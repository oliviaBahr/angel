use crate::output::stdout;

pub const VERSION: &str = "0.1.0";

pub fn run() {
    stdout::writeln(&format!("angel {}", VERSION));
}
