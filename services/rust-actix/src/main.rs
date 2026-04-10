mod storage;
mod file_storage;
#[cfg(test)]
mod storage_test;

use actix_web::{web, App, HttpServer, HttpResponse, middleware};
use chrono::Utc;
use serde_json::json;
use std::sync::Arc;

use file_storage::FileLogStorage;
use storage::LogStorage;

const LOG_FILE: &str = "/data/log.txt";

/// Приложение с зависимостями (Dependency Injection)
struct AppState {
    storage: Arc<dyn LogStorage>,
}

fn format_log_line(hostname: &str, timestamp: &str) -> String {
    format!("[{}] {}\n", hostname, timestamp)
}

async fn write_time(data: web::Data<AppState>) -> HttpResponse {
    let now = Utc::now().to_rfc3339();
    let hostname = std::env::var("HOSTNAME").unwrap_or_else(|_| "unknown".to_string());
    let line = format_log_line(&hostname, &now);

    match data.storage.write_line(&line) {
        Ok(()) => HttpResponse::Ok().json(json!({"message": "time written", "time": now})),
        Err(e) => HttpResponse::InternalServerError().json(json!({"error": e.to_string()})),
    }
}

async fn read_log(data: web::Data<AppState>) -> HttpResponse {
    match data.storage.read_all() {
        Ok(content) => HttpResponse::Ok().json(json!({"content": content})),
        Err(e) => HttpResponse::InternalServerError().json(json!({"error": e.to_string()})),
    }
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
    // Инициализация зависимостей (Dependency Injection)
    let storage = FileLogStorage::new(LOG_FILE);
    let app_state = web::Data::new(AppState {
        storage: Arc::new(storage),
    });

    HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
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
        let resp = root().await;
        assert!(resp.status().is_success());

        let body = resp.into_body().try_into_bytes().unwrap();
        let json: Value = serde_json::from_slice(&body).unwrap();

        assert_eq!(json["service"], "rust-actix");
        assert_eq!(json["status"], "running");
    }

    #[actix_web::test]
    async fn test_health_response() {
        let resp = health().await;
        assert!(resp.status().is_success());

        let body = resp.into_body().try_into_bytes().unwrap();
        let json: Value = serde_json::from_slice(&body).unwrap();

        assert_eq!(json["status"], "healthy");
    }
}
