# XXX: source config for theming and common config
# Source ../../config.tape

# Output
Output gnoland-start-lazy.gif

# Requirement
Require gnoland

# Set up, should be in config.
Set FontSize 18
Set Width 1200
Set Height 800

## Run

Type "gnoland start --lazy"
Sleep 500ms
Enter

# Wait a bit and Exit
Sleep 5s
Ctrl+c