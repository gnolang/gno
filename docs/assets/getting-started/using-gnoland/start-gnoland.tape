# XXX: source config for theming and common config
# Source ../../config.tape

# Output
Output gnokey-generate.gif

# Requirement
Require gnoland

# Set up, should be in config.
Set FontSize 18
Set Width 1200
Set Height 800

## Run

Type "gnoland start"
Sleep 500ms
Enter

# Wait a bit and Exit
Sleep 10s
Ctrl+c