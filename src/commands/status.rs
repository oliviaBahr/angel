use crate::angel::Angel;
use crate::cli::NameArgs;
use crate::display;
use crate::error::Result;
use crate::launchctl;
use crate::output::stdout;

pub fn run(angel: &Angel, args: &NameArgs) -> Result<()> {
    let daemon = angel.daemons.get_match(&args.name, args.exact)?;
    let result = launchctl::print(daemon)?;

    match result.success() {
        false => stdout::error(&format!(
            "failed to get status for {}: {}",
            daemon.name, result.stderr
        )),
        true => {}
    }

    // Parse basic info from launchctl print output
    // For now, just show the raw output since parser is not implemented
    let status = extract_status(&result.output);

    stdout::writeln(&display::bold(&daemon.name));
    stdout::writeln(&display::format_status_dot(&status));

    let mut table = display::create_table();
    table.add_row(vec!["Domain:".to_string(), daemon.domain_str()]);
    table.add_row(vec![
        "Source:".to_string(),
        daemon
            .source_path
            .as_ref()
            .map_or("-".to_string(), |p| p.display().to_string()),
    ]);

    match args.verbose {
        true => {
            // Add plist fields if available
            match &daemon.plist {
                Some(plist) => {
                    match &plist.program {
                        Some(program) => {
                            table.add_row(vec!["Program:".to_string(), program.clone()]);
                        }
                        None => {}
                    }
                    match &plist.program_arguments {
                        Some(program_args) => {
                            table.add_row(vec![
                                "ProgramArguments:".to_string(),
                                program_args.join(" "),
                            ]);
                        }
                        None => {}
                    }
                    // Add more fields as needed
                }
                None => {}
            }
        }
        false => {}
    }

    stdout::writeln(&table);
    Ok(())
}

fn extract_status(output: &str) -> String {
    // Simple extraction - look for common status indicators
    match () {
        _ if output.contains("state = running") => "running".to_string(),
        _ if output.contains("state = stopped") => "stopped".to_string(),
        _ if output.contains("state = launched") => "launched".to_string(),
        _ if output.contains("state = exited") => "exited".to_string(),
        _ => "unknown".to_string(),
    }
}
