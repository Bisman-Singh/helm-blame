Output demo.gif

Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Catppuccin Mocha"
Set Padding 20

Type "# Simple chart with production overrides"
Enter
Sleep 500ms

Type "helm blame examples/simple -f examples/simple/production.yaml --show-shadowed"
Enter
Sleep 3s

Type ""
Enter
Sleep 500ms

Type "# Umbrella chart with subcharts + --set override"
Enter
Sleep 500ms

Type "helm blame examples/umbrella -f examples/umbrella/staging.yaml --set app.image.tag=v1.2.0 --show-shadowed"
Enter
Sleep 4s

Type ""
Enter
Sleep 500ms

Type "# JSON output for CI pipelines"
Enter
Sleep 500ms

Type "helm blame examples/simple -f examples/simple/production.yaml -o json | head -20"
Enter
Sleep 3s
