use comfy_table::Table;
use nu_ansi_term::{Color, Style};

pub fn bold(text: &str) -> String {
    Style::new().bold().paint(text).to_string()
}

pub fn format_status_dot(status: &str) -> String {
    let (color, dot) = match status {
        "running" => (Color::Green, "●"),
        "stopped" => (Color::Red, "●"),
        "launched" => (Color::Yellow, "●"),
        "exited" => (Color::Blue, "●"),
        _ => (Color::Magenta, "●"),
    };
    format!("{} {}", color.paint(dot), status)
}

pub fn create_table() -> Table {
    let mut table = Table::new();
    table.load_preset(comfy_table::presets::UTF8_BORDERS_ONLY);
    table.apply_modifier(comfy_table::modifiers::UTF8_ROUND_CORNERS);
    table
}
