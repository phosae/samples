use std::{env, io, time::Duration};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let url = match env::args().nth(1) {
        Some(url) => url,
        None => {
            println!("Usage: {} <url>", env::args().nth(0).expect("not empty"));
            return Ok(());
        }
    };

    let max_retries = 3;
    let mut retry_count = 1;
    let mut delay = Duration::from_millis(500);

    while retry_count <= max_retries {
        let ret = reqwest::blocking::get(&url);

        if should_retry(&ret) {
            println!("retry on err: #{:#?}, round #{:#?}", ret, retry_count);
            retry_count += 1;
            std::thread::sleep(delay);
            delay *= 2; // exponential backoff
            continue;
        }

        println!("Got final result");
        match ret {
            Ok(resp) => {
                println!("Got response: {:#?}", resp);
            }
            Err(err) => {
                println!("Got unrecoverable Err: {:#?}", err);
            }
        }
        break;
    }

    Ok(())
}

fn should_retry(ret: &Result<reqwest::blocking::Response, reqwest::Error>) -> bool {
    match ret {
        Ok(resp) => {
            return resp.status().as_u16() >= 500
                || resp.status() == http::StatusCode::TOO_MANY_REQUESTS
        }
        Err(err) => {
            if err.is_connect() || err.is_timeout() {
                return true;
            }
            if let Some(err) = get_source_error_type::<hyper::Error>(&err) {
                // The hyper::Error(IncompleteMessage) is raised if the HTTP response is well formatted but does not contain all the bytes.
                // This can happen when the server has started sending back the response but the connection is cut halfway thorugh.
                // We can safely retry the call, hence marking this error as [`Retryable::Transient`].
                // Instead hyper::Error(Canceled) is raised when the connection is
                // gracefully closed on the server side.
                if err.is_incomplete_message() || err.is_canceled() {
                    return true;
                } else {
                    if let Some(err) = get_source_error_type::<io::Error>(err) {
                        return err.kind() == std::io::ErrorKind::ConnectionReset
                            || err.kind() == std::io::ErrorKind::ConnectionAborted;
                    }
                }
            }
            return false;
        }
    }
}

fn get_source_error_type<T: std::error::Error + 'static>(
    err: &dyn std::error::Error,
) -> Option<&T> {
    let mut source = err.source();

    while let Some(err) = source {
        if let Some(err) = err.downcast_ref::<T>() {
            return Some(err);
        }
        source = err.source();
    }
    None
}
