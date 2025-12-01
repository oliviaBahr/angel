use crate::angel::Angel;
use crate::cli::NameArgs;
use crate::error::Result;
use crate::launchctl;

pub fn run(angel: &Angel, args: &NameArgs) -> Result<()> {
    let daemon = angel.daemons.get_match(&args.name, args.exact);
    let output = launchctl::enable(daemon)?;
    print!("{}", output);
    println!("enabled {}", daemon.name);
    Ok(())
}


