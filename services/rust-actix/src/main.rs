use actix_web::{web, App, HttpServer, HttpResponse, middleware};
use chrono::Utc;
use nix::fcntl::{flock, FlockArg};
use serde_json::json;
use std::fs::OpenOptions;
use std::io::{Read, Write};
use std::os::unix::io::AsRawFd;

const LOG_FILE: &str = "/data/log.txt";

async fn write_time() -> HttpResponse {
    let now = Utc::now().to_rfc3339();
    let hostname = std::env::var("HOSTNAME").unwrap_or_else(|_| "unknown".to_string());
    let line = format!("[{}] {}\n", hostname, now);

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
    HttpResponse::Ok().json(json!({
        "service": "rust-actix",
        "status": "running"
    }))
}

async fn health() -> HttpResponse {
    HttpResponse::Ok().json(json!({
        "status": "healthy"
    }))
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
