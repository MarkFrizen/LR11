use actix_web::{web, App, HttpServer, HttpResponse, middleware};
use chrono::Utc;
use nix::fcntl::{flock, FlockArg};
use serde_json::json;
use std::fs::OpenOptions;
use std::io::{Read, Write};
use std::os::unix::io::AsRawFd;

const LOG_FILE: &str = "/data/log.txt";

fn format_log_line(hostname: &str, timestamp: &str) -> String {
    format!("[{}] {}\n", hostname, timestamp)
}

fn create_root_response() -> HttpResponse {
    HttpResponse::Ok().json(json!({
        "service": "rust-actix",
        "status": "running"
    }))
}

fn create_health_response() -> HttpResponse {
    HttpResponse::Ok().json(json!({
        "status": "healthy"
    }))
}

async fn write_time() -> HttpResponse {
    let now = Utc::now().to_rfc3339();
    let hostname = std::env::var("HOSTNAME").unwrap_or_else(|_| "unknown".to_string());
    let line = format_log_line(&hostname, &now);

    match OpenOptions::new()
        .create(true)
        .append(true)
        .open(LOG_FILE)
    {
        Ok(f) => {
            let fd = f.as_raw_fd();
            if let Err(e) = flock(fd, FlockArg::LockExclusive) {
                return HttpResponse::InternalServerError().json(json!({"error": e.to_string()}));
            }
            let mut f = f;
            if let Err(e) = f.write_all(line.as_bytes()) {
                let _ = flock(fd, FlockArg::Unlock);
                return HttpResponse::InternalServerError().json(json!({"error": e.to_string()}));
            }
            let _ = flock(fd, FlockArg::Unlock);
            HttpResponse::Ok().json(json!({"message": "time written", "time": now}))
        }
        Err(e) => HttpResponse::InternalServerError().json(json!({"error": e.to_string()})),
    }
}

async fn read_log() -> HttpResponse {
    let mut file = match OpenOptions::new().read(true).open(LOG_FILE) {
        Ok(f) => f,
        Err(_) => {
            return HttpResponse::Ok().json(json!({"content": ""}));
        }
    };

    let fd = file.as_raw_fd();
    if let Err(e) = flock(fd, FlockArg::LockShared) {
        return HttpResponse::InternalServerError().json(json!({"error": e.to_string()}));
    }

    let mut content = String::new();
    if let Err(e) = file.read_to_string(&mut content) {
        let _ = flock(fd, FlockArg::Unlock);
        return HttpResponse::InternalServerError().json(json!({"error": e.to_string()}));
    }

    let _ = flock(fd, FlockArg::Unlock);
    HttpResponse::Ok().json(json!({"content": content}))
}

async fn root() -> HttpResponse {
    create_root_response()
}

async fn health() -> HttpResponse {
    create_health_response()
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .wrap(middleware::Logger::default())
            .route("/", web::get().to(root))
            .route("/health", web::get().to(health))
            .route("/write", web::get().to(write_time))
            .route("/read", web::get().to(read_log))
    })
    .bind("0.0.0.0:8082")?
    .run()
    .await
}

#[cfg(test)]
mod tests {
    use super::*;
    use actix_web::body::MessageBody;
    use serde_json::Value;

    #[test]
    fn test_format_log_line() {
        let result = format_log_line("test-host", "2026-04-09T12:00:00+00:00");
        let expected = "[test-host] 2026-04-09T12:00:00+00:00\n";
        assert_eq!(result, expected);
    }

    #[test]
    fn test_format_log_line_empty_hostname() {
        let result = format_log_line("", "2026-04-09T12:00:00+00:00");
        let expected = "[] 2026-04-09T12:00:00+00:00\n";
        assert_eq!(result, expected);
    }

    #[test]
    fn test_format_log_line_with_special_chars() {
        let result = format_log_line("host:8080", "time");
        let expected = "[host:8080] time\n";
        assert_eq!(result, expected);
    }

    #[actix_web::test]
    async fn test_root_response() {
        let resp = create_root_response();
        assert!(resp.status().is_success());

        let body = resp.into_body().try_into_bytes().unwrap();
        let json: Value = serde_json::from_slice(&body).unwrap();
        
        assert_eq!(json["service"], "rust-actix");
        assert_eq!(json["status"], "running");
    }

    #[actix_web::test]
    async fn test_health_response() {
        let resp = create_health_response();
        assert!(resp.status().is_success());

        let body = resp.into_body().try_into_bytes().unwrap();
        let json: Value = serde_json::from_slice(&body).unwrap();
        
        assert_eq!(json["status"], "healthy");
    }
}
