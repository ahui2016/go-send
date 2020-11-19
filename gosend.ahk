#S::
ClipSaved := clipboard  ; Convert any copied files, HTML, or other formatted text to plain text.
ClipSaved := Trim(ClipSaved)
Run, go-send-cli.exe -text "%ClipSaved%"
return

#F::
ClipSaved := clipboard  ; Convert any copied files, HTML, or other formatted text to plain text.
ClipSaved := Trim(ClipSaved)
Run, go-send-cli.exe -file "%ClipSaved%"
return
