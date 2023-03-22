use nix::libc;
use nix::sys::wait::waitpid;
use nix::unistd::fork;
use nix::unistd::ForkResult::{Child, Parent};
use std::process;
use std::thread::sleep;
use std::time;

fn main() {
    let delay = time::Duration::from_secs(1);
    let reap: bool = std::env::var("REAP").unwrap_or("false".to_string()) == "true";

    let pid = process::id();
    println!("{}", pid);

    for i in 1..=60 {
        println!("{} . {}", pid, i);
        match unsafe { fork() } {
            Ok(Parent { child, .. }) => {
                println!(
                    "Continuing execution in parent process, fork child: {}",
                    child
                );
                if reap {
                    waitpid(child, None).unwrap();
                }
            }
            Ok(Child) => {
                unsafe { libc::_exit(0) };
            }
            Err(_) => println!("Fork failed"),
        }
        sleep(delay);
    }
}
