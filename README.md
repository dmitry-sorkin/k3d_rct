# K3D Retraction calibration towers generator

Simple retraction towers generator for retraction calibration written in golang. Hosted [>here<](https://k3d.tech/calibrations/retractions/rct.html?lang=en). 
The author is not a professional programmer. The code is terrible. When reading the code, avoid psychological trauma.

# Building

Install golang and then simply run build.bat/build.sh, it should generate WASM file.

⚠️WebAssembly files will not work from locally opened html. You need to use any web server to run it. For example, simple python web server: `python -m http.server 8080`

------

## TODO

- [X] English localization
- [X] Replace LA logic with firmware selector and value
- [X] Implement flow setting
- [X] Implement start/end G-Code setting
- [X] Implement reset to default settings button
- [ ] Change validating logic so that the values are checked before generating file

[Русская версия](README_RU.md)
