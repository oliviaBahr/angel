use crate::angel::Angel;
use crate::cli::NameArgs;
use crate::display;
use crate::error::Result;
use crate::launchctl;
use crate::output::{is_verbose, stdout};
use nu_ansi_term::Color;

pub fn run(angel: &Angel, args: &NameArgs) -> Result<()> {
    let daemon = angel.daemons.get_match(&args.name, args.exact)?;
    let result = launchctl::print(daemon)?;

    let status = extract_status(&result.output);
    let color = if result.success() { None } else { Some(Color::Red) };

    stdout::writeln(&display::bold(&daemon.name));
    stdout::writeln(&display::format_status_dot(&status, color));

    let mut table = display::create_table();
    table.add_row(vec!["Domain:".to_string(), daemon.domain_str()]);
    table.add_row(vec![
        "Source:".to_string(),
        daemon.source_path.as_ref().map_or("-".to_string(), |p| p.display().to_string()),
    ]);

    if is_verbose() {
        // Add plist fields if available
        if let Some(plist) = &daemon.plist {
            if let Some(program) = &plist.program {
                table.add_row(vec!["Program:".to_string(), program.clone()]);
            }
            if let Some(program_arguments) = &plist.program_arguments {
                table.add_row(vec!["ProgramArguments:".to_string(), program_arguments.join(" ")]);
            }
        }
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
