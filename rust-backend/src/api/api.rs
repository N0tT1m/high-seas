pub use once_cell::sync::Lazy;
pub use std::fs::{File, OpenOptions};
use tokio::sync::Mutex;

// Make the static variable public
pub static LOG_FILE: Lazy<Mutex<File>> = Lazy::new(|| {
    let file = OpenOptions::new()
        .create(true)
        .append(true)
        .open("api.log")
        .expect("Failed to create log file");
    Mutex::new(file)
});

