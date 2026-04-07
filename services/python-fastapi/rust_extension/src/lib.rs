use pyo3::prelude::*;

/// Rust-функция: вычисляет n-е число Фибоначчи
#[pyfunction]
fn fibonacci(n: u64) -> u64 {
    match n {
        0 => 0,
        1 => 1,
        _ => {
            let (mut a, mut b) = (0, 1);
            for _ in 2..=n {
                let temp = a + b;
                a = b;
                b = temp;
            }
            b
        }
    }
}

/// Rust-функция: приветствие (проверка вызова из Python)
#[pyfunction]
fn rust_hello(name: &str) -> String {
    format!("Hello from Rust, {}!", name)
}

/// Python-модуль из Rust
#[pymodule]
fn rust_ext(m: &Bound<'_, PyModule>) -> PyResult<()> {
    m.add_function(wrap_pyfunction!(fibonacci, m)?)?;
    m.add_function(wrap_pyfunction!(rust_hello, m)?)?;
    Ok(())
}
