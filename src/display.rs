use crate::types::{Daemon, Domain};
use comfy_table::Table;
use nu_ansi_term::Color;
use std::path::{Path, PathBuf};

pub fn format_status_dot(status: &str, color: Option<Color>) -> String {
    let (color, dot) = match status {
        "running" => (Color::Green, "●"),
        "not running" => (Color::Default, "●"),
        "stopped" => (Color::Red, "●"),
        "launched" => (Color::Yellow, "●"),
        "exited" => (Color::Blue, "●"),
        _ => (color.unwrap_or(Color::Magenta), "●"),
    };
    format!("{} {}", color.paint(dot), status)
}

pub fn color_domain(domain: &Domain) -> String {
    match domain {
        Domain::System => Color::LightPurple.paint(domain.to_string()).to_string(),
        Domain::User(_) => Color::LightGreen.paint(domain.to_string()).to_string(),
        Domain::Gui(_) => Color::Cyan.paint(domain.to_string()).to_string(),
        Domain::Unknown => domain.to_string(),
    }
}

pub fn create_table() -> Table {
    let mut table = Table::new();
    table.load_preset(comfy_table::presets::UTF8_BORDERS_ONLY);
    table.apply_modifier(comfy_table::modifiers::UTF8_ROUND_CORNERS);
    table
}

fn compress_path(path: &Path) -> String {
    if let Some(home) = dirs::home_dir() {
        if let Ok(relative) = path.strip_prefix(&home) {
            return format!("~/{}", relative.display());
        }
    }
    path.display().to_string()
}

pub fn display_path(daemon: &Daemon, long: bool) -> String {
    match &daemon.source_path {
        Some(path) => match path.is_symlink() {
            true => match long {
                true => format!(
                    "{} → {}",
                    compress_path(path),
                    compress_path(&path.read_link().unwrap_or(PathBuf::from("")))
                ),
                false => format!("  {}", compress_path(path)),
            },
            false => compress_path(path),
        },
        None => "-".to_string(),
    }
}
