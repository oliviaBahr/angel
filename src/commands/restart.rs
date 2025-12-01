use crate::angel::Angel;
use crate::cli::NameArgs;
use crate::error::Result;
use crate::launchctl;

pub fn run(angel: &Angel, args: &NameArgs) -> Result<()> {
    let daemon = angel.daemons.get_match(&args.name, args.exact);
    let output = launchctl::kickstart_kill(daemon)?;
    print!("{}", output);
    println!("restarted {}", daemon.name);
    Ok(())
}
