use actix_web::{web, App, HttpServer, HttpResponse, middleware};
use serde_json::json;

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
    })
    .bind("0.0.0.0:8080")?
    .run()
    .await
}
