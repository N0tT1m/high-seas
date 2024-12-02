mod internal;

use std::sync::Arc;
use axum::{
    routing::post,
    Router,
};
use tower_http::cors::{Any, CorsLayer};
use log::error;
use tracing_subscriber::{filter::LevelFilter, fmt::{self, format::FmtSpan}, prelude::*, EnvFilter};
use tracing_appender::rolling::{RollingFileAppender, Rotation};
use std::net::SocketAddr;
use axum::http::Method;
use axum::routing::get;
use axum_server::tls_rustls::RustlsConfig;

pub fn setup_logging() {
    // Create file appender
    let file_appender = RollingFileAppender::new(
        Rotation::DAILY,
        "logs",
        "high-seas.log",
    );

    // Create an EnvFilter that enables everything
    let filter = EnvFilter::new("trace,hyper=debug,tower_http=debug")
        .add_directive(LevelFilter::TRACE.into());

    // Create the subscriber
    let subscriber = tracing_subscriber::registry()
        .with(fmt::Layer::new()
            .with_file(true)
            .with_line_number(true)
            .with_thread_ids(true)
            .with_thread_names(true)
            .with_span_events(FmtSpan::FULL)
            .with_writer(file_appender)
            .with_ansi(false)
            .with_target(true)
            .with_level(true)
            .with_filter(filter)
        );

    // Set the subscriber as the default
    if let Err(e) = tracing::subscriber::set_global_default(subscriber) {
        eprintln!("Failed to set up logging: {}", e);
    }

    // Log initial startup
    error!("Logging system initialized with TRACE level enabled");
}

#[tokio::main]
async fn main() {
    // Initialize logging first
    setup_logging();

    // Start your application...
    tracing::info!("Starting application...");

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods([Method::POST])
        .allow_headers(Any);

    let app = Router::new()
        .route("/movie/query", post(internal::api::query_movie_request))

        .layer(cors);

    // let listener = tokio::net::TcpListener::bind("localhost:8782")
    //     .await
    //     .unwrap();
    // tracing::debug!("listening on {}", listener.local_addr().unwrap());
    // axum::serve(listener, app).await.unwrap();

    // Configure the domain and certificate
    let config = RustlsConfig::from_pem_file(
        "/usr/local/bin/fullchain.pem",
        "/usr/local/bin/privkey.pem",
    )
        .await
        .expect("Failed to load certificate and private key");

    // Run the server
    let addr = SocketAddr::from(([0, 0, 0, 0], 8877));
    axum_server::bind_rustls(addr, config)
        .serve(app.into_make_service())
        .await
        .expect("Failed to start server");

    // Run the server

}