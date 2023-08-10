const calibrator_version = 'v1.8';
window.calibrator_version = calibrator_version;
var savedSegmentsInfo = null;

function download(filename, text) {
    var element = document.createElement('a');
    element.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(text));
    element.setAttribute('download', filename);

    element.style.display = 'none';
    document.body.appendChild(element);

    element.click();

    document.body.removeChild(element);
}

function saveTextAsFile(filename, text) {
    var textFileAsBlob = new Blob([text], { type: 'text/plain' });

    var downloadLink = document.createElement("a");
    downloadLink.download = filename;
    if (window.webkitURL != null) {
        // Chrome allows the link to be clicked without actually adding it to the DOM.
        downloadLink.href = window.webkitURL.createObjectURL(textFileAsBlob);
    } else {
        // Firefox requires the link to be added to the DOM before it can be clicked.
        downloadLink.href = window.URL.createObjectURL(textFileAsBlob);
        downloadLink.onclick = destroyClickedElement;
        downloadLink.style.display = "none";
        document.body.appendChild(downloadLink);
    }

    downloadLink.click();
}

function showError(value) {
    var container = document.getElementById("resultContainer");
    var output = document.createElement("textarea");
    output.id = "gCode";
    // output.name = "gCode";
    output.cols = "80";
    output.rows = "10";
    output.value = value;
    // output.className = "css-class-name"; // set the CSS class
    container.appendChild(output); //appendChild
}

function destroyClickedElement(event) {
    // remove the link from the DOM
    document.body.removeChild(event.target);
}

var formFields = [
    "bedX",
    "bedY",
    "zOffset",
    "delta",
    "bedProbe",
    "hotendTemperature",
    "bedTemperature",
    "cooling",
    "lineWidth",
    "firstLayerLineWidth",
    "layerHeight",
    "printSpeed",
    "firstLayerPrintSpeed",
    "travelSpeed",
    "initRetractLength",
    "endRetractLength",
    "initRetractSpeed",
    "endRetractSpeed",
    "numSegments",
    "segmentHeight",
    "kFactor2",
    "towerSpacing",
    "flow",
    "firmwareMarlin",
    "firmwareKlipper",
    "firmwareRRF",
	"hardmode",
	"startGcode",
	"endGcode",
];
var segmentFields = [
    "initRetractLength",
    "endRetractLength",
    "initRetractSpeed",
    "endRetractSpeed",
    "numSegments",
];
var segmentKeys = [
	"init_retract_length",
	"end_retract_length",
	"init_retract_speed",
	"end_retract_speed",
	"num_segments",
];

var saveForm = function () {
    for (var elementId of formFields) {
        var element = document.getElementById(elementId);
        if (element) {
            var saveValue = element.value;
            if (elementId == 'delta' || elementId == 'bedProbe' || elementId == 'firmwareMarlin' || elementId == 'firmwareKlipper' || elementId == 'firmwareRRF' || elementId == 'hardmode') {
                saveValue = element.checked;
            }
            localStorage.setItem(elementId, saveValue);
        }
    }
}

function loadForm() {
    for (var elementId of formFields) {
        let loadValue = localStorage.getItem(elementId);
        if (loadValue === undefined) {
            continue;
        }

        var element = document.getElementById(elementId);
        if (element) {
            if (elementId == 'delta' || elementId == 'bedProbe' || elementId == 'firmwareMarlin' || elementId == 'firmwareKlipper' || elementId == 'firmwareRRF' || elementId == 'hardmode') {
				element.checked = loadValue == 'true';
            } else {
                if (loadValue != null) {
                    element.value = loadValue;
                }
            }
            
        }
    }
}

function initForm() {
	loadForm();
    for (var elementId of formFields) {
        var element = document.getElementById(elementId);
        element.addEventListener('change', function(e) {
			saveForm();
			
			var el = e.target;
			var id = el.id;
			
			if (segmentFields.indexOf(id) != -1) {
				checkSegments();
			} else {
				checkGo();
			}
		});
    }
	for (var elementId of segmentFields) {
		var element = document.getElementById(elementId);
		element.addEventListener('focusin', function(e) {
			checkSegments();
		});
		element.addEventListener('focusout', function(e) {
			if (e.relatedTarget == undefined || segmentFields.indexOf(e.relatedTarget.id) == -1) {
				checkGo();
			}
		});
	}
}

function initLang(key) {
	var values = window.lang.values;
	switch (key) {
		case 'de':
			values['header.title'] = 'K3D Einzug-Kalibrator';
			values['header.description'] = 'Detaillierte Beschreibung der Funktion finden Sie <a href="http://k3d.tech/calibrations/retractions/">hier.</a>';
			values['header.move_exceeds'] = 'Falls Sie die Fehlermeldung "Move exceeds maximum extrusion" bekommen, dann lesen Sie <a href="http://k3d.tech/calibrations/retractions/#move-exceeds-maximum-extrusion">hier</a>';
			values['header.language'] = 'Sprache: ';
			
			values['table.header.parameter'] = 'Parameter';
			values['table.header.value'] = 'Wert';
			values['table.header.description'] = 'Beschreibung';
			
			values['table.bed_size_x.title'] = 'Druckbettgröße an der X-Achse';
			values['table.bed_size_x.description'] = '[mm] Für kartesische Drucker - Maximalkoordinate an der X-Achse<br>Für Delta-Drucker - <b>Durchmesser des Druckbettes</b>';
			values['table.bed_size_y.title'] = 'Druckbettgröße an der Y-Achse';
			values['table.bed_size_y.description'] = '[mm] Für kartesische Drucker - Maximalkoordinate an der Y-Achse<br>Für Delta-Drucker - <b>Durchmesser des Druckbettes</b>';
			values['table.z_offset.title'] = 'Z-offset';
			values['table.z_offset.description'] = '[mm] Vertikale verschiebung der Türmchen. Damit kann zu dünne oder zu dicke erste Schicht ausgeglichen werden. Allgemein lassen Sie es bei null';
			values['table.delta.title'] = 'Nullpunkt in der Mitte des Druckbettes';
			values['table.delta.description'] = 'Für kartesische Drucker muss abgeschaltet sein, für Delta - eingeschaltet';
			values['table.bed_probe.title'] = 'Autokalibrierung';
			values['table.bed_probe.description'] = 'Soll vor dem Drucken Autokalibrierung (G29) durchgeführt werden? Lassen Sie es ausgeschaltet, falls nicht gewünscht oder kein Sensor (BL Tuch o.ä.) vorhanden.';
			values['table.hotend_temp.title'] = 'Drucktemperatur';
			values['table.hotend_temp.description'] = '[°C] Temperatur des Extruders (Hotend). die beim Drucken für Ihr Filament benötigt wird.';
			values['table.bed_temp.title'] = 'Druckbetttemperatur';
			values['table.bed_temp.description'] = '[°C] Temperatur, die vor dem Drucken eingestellt wird. Vor der Nullpunktfahrt und der Autokalibrierung wartet der Drucker, bis der eingestellte Wert erreicht wird.';
			values['table.flow.title'] = 'Fluss';
			values['table.flow.description'] = '[%] Fluss in Prozent. Wird für die Kompensation der Über- oder Unterextrusion benötigt.';
			values['table.fan_speed.title'] = 'Lüftergeschwindigkeit';
			values['table.fan_speed.description'] = '[%] Drehzahl des Lüfters in Prozent. Um den Temperatureinfall zu vermeiden, wird der Lüfter bei der ersten Schicht mit 1/3 vom eingestellten Wert, bei der 2. mit 2/3, bei der 4. mit dem eingestellten Wert.';
			values['table.line_width.title'] = 'Linienbreite';
			values['table.line_width.description'] = '[mm] Linienbreite, mit der die Türmchen gedruckt werden. Allgemein wie Düsendurchmesser.';
			values['table.first_line_width.title'] = 'Linienbreite der ersten Schicht';
			values['table.first_line_width.description'] = '[mm] Linienbreite, mit der die Unterlage für die Türmchen gedruckt wird. Allgemein wird 150% des Düsendurchmesser empfohlen.';
			values['table.layer_height.title'] = 'Schichtdicke';
			values['table.layer_height.description'] = '[mm] Höhe der einzelnen Schichten. Allgemein 50% vom Düsendurchmesser.';
			values['table.print_speed.title'] = 'Druckgeschwindigkeit';
			values['table.print_speed.description'] = '[mm/s] Geschwindigkeit, mit der die Türmchen gedruckt werden.';
			values['table.first_print_speed.title'] = 'Druckgeschwindigkeit der ersten Schicht.';
			values['table.first_print_speed.description'] = '[mm/s] Geschwindigkeit, mit der die Untelage gedruckt wird.';
			values['table.travel_speed.title'] = 'Bewegungsgeschwindigkeit';
			values['table.travel_speed.description'] = '[mm/s] Geschwindigkeit zwischen der Türmchen.';
			values['table.init_retract_length.title'] = 'Anfangs-Einzug';
			values['table.init_retract_length.description'] = '[mm] Einzugslänge, die für das unterste Segment angewendet wird.';
			values['table.end_retract_length.title'] = 'End-Einzug';
			values['table.end_retract_length.description'] = '[mm] Einzugslänge, die für das oberste Segment angewendet wird. Zwischen dem untersten und dem obersten Segment ändert sich die Einzugslänge stufenweise. Tragen Sie den Wert der Anfangs-Einzugslänge, wenn diese sich nicht ändern soll - so können Sie nur die Einzugsgeschwindigkeit kalibrieren.';
			values['table.init_retract_speed.title'] = 'Anfangs-Einzugsgeschwindigkeit';
			values['table.init_retract_speed.description'] = '[mm/s] Einzugsgeschwindigkeit, die am untersten Segment angewendet wird.';
			values['table.end_retract_speed.title'] = 'End-Einzugsgeschwindigkeit';
			values['table.end_retract_speed.description'] = '[mm/s] Einzugsgeschwindigkeit für das oberste Segment. Zwischen dem untersten und dem obersten Segmente ändert sich die Einzugsgeschwindigkeit stufenweise. Tragen Sie die Anfangs-Einzugsgeschwindigkeit, wenn keine Änderung gewünscht ist - so können Sie nur die Einzugslänge kalibrieren.';
			values['table.num_segments.title'] = 'Anzahl der Segmente';
			values['table.num_segments.description'] = 'Anzahl der Segmente im Türmchen. Innerhalb eines jeden Segmentes bleiben Einzugslänge und Einzugsgeschwindigkeit konstant. Einzelne Segmente sind visuell getrennt, um die Analyse des Modells zu vereinfachen,';
			values['table.segment_height.title'] = 'Segmenthöhe';
			values['table.segment_height.description'] = '[mm] Höhe der einzelnen Segmente der Türmchen. Zum Beispiel: Wenn Segmenthöhe 3mm und Anzahl der Segmente 10 ist, beträgt die Gesamthöhe der Türmchen 30mm.';
			values['table.k_factor.title'] = 'Linear Advance (K-Faktor)';
			values['table.k_factor.description'] = 'Tragen Sie Ihren Wert für Linear/Pressure Advance. Wenn Sie es nicht benutzen oder nicht wissen, was Linear/Pressure Advance bedeutet, lassen Sie es bei null.';
			values['table.tower_spacing.title'] = 'Abstand zwischen den Türmchen';
			values['table.tower_spacing.description'] = '[mm] In der Regel reichen 100mm. Für großflächige Drucker, die oft große Modelle drucken, wird ungefähr die Hälfte der längeren Seite des Druckbettes empfohlen.';
			values['table.firmware.title'] = 'Firmware';
			values['table.firmware.description'] = 'Firmware, die in Ihrem Drucker installiert ist. Wenn nicht bekannt, dann, wahrschainlich Marlin';
			values['table.start_gcode.title'] = 'Start G-Code';
			values['table.start_gcode.description'] = 'G-Code, welches vor dem Drucken ausgeführt wird. Änderungen auf eigenes Risiko! Keine Haftung bei Schäden! Liste möglicher Platzhalter:<br><b>$LA</b> - Kompletter Installationsbefehl Faktor LA/PA<br><b>$BEDTEMP</b> - Druckbetttemperatur<br><b>$HOTTEMP</b> - Extrudertemperatur<br><b>$G29</b> - Befehl zum Abtasten des Druckbettes<br><b>$FLOW</b> - Fluss';
			values['table.end_gcode.title'] = 'End G-Code';
			values['table.end_gcode.description'] = 'G-Code, welches nach dem Drucken ausgeführt wird. Änderungen auf eigenes Risiko! Keine Haftung bei Schäden!';
			values['table.hardmode.title'] = 'harter Modus';
			values['table.hardmode.description'] = 'Im Normalmodus (der Parameter ist deaktiviert) wird die Druckreihenfolge des Turms mit Optimierungen wie in einem Slicer generiert, im erweiterten Modus durch eine nicht optimale Methode. Es wird empfohlen, es nur dann einzuschalten, wenn der Normalmodus zu optimistische Ergebnisse liefert. Lesen Sie mehr in der Anleitung'
			
			values['generator.generate_and_download'] = 'Generieren und Herunterladen';		
			values['generator.generate_button_loading'] = 'G-Code wird generiert...';		
			values['generator.segment'] = ';Segment %d:   %smm @ %smm/s\n';
			values['generator.reset_to_default'] = 'Einstellungen zurücksetzen';
			
			values['navbar.back'] = ' Zurück ';
			values['navbar.site'] = 'Webseite';
			
			values['error.bed_size_x.format'] = 'X-Achsen-Abmessung - Format Fehler';
			values['error.bed_size_x.small_or_big'] = 'Die X-Achsenlänge des Druckbettes ist falsch (weniger als 100 oder mehr als 1000 mm)';
			values['error.bed_size_y.format'] = 'Y-Achsen-Abmessung - Format Fehler';
			values['error.bed_size_y.small_or_big'] = 'Die Y-Achsenlänge des Druckbettes ist falsch (weniger als 100 oder mehr als 1000 mm)';
			values['error.hotend_temp.format'] = 'Extrudertemperatur - Format Fehler';
			values['error.hotend_temp.too_low'] = 'Temperatur des Extruders ist zu niedrig';
			values['error.hotend_temp.too_high'] = 'Temperatur des Extruders ist zu hoch';
			values['error.bed_temp.format'] = 'Druckbetttemperatur - Format Fehler: ';
			values['error.bed_temp.too_high'] = 'Druckbetttemperatur ist zu hoch';
			values['error.fan_speed.format'] = 'Lüftergeschwindigkeit - Format Fehler';
			values['error.line_width.format'] = 'Linienbreite - Format Fehler';
			values['error.line_width.small_or_big'] = 'Linienbreite ist falsch (weniger als 0.1 oder mehr als 2.0 mm)';
			values['error.first_line_width.format'] = 'Linienbreite der ersten Schicht - Format Fehler';
			values['error.first_line_width.small_or_big'] = 'Linienbreite der ersten Schicht ist falsch (weniger als 0.1 oder mehr als 2.0 mm)';
			values['error.layer_height.format'] = 'Schichtdicke - Format Fehler';
			values['error.layer_height.small_or_big'] = 'Schichtdicke ist falsch (weniger als 0.05 mm oder mehr als 75% der Linienbreite)';
			values['error.print_speed.format'] = 'Druckgeschwindigkeit - Format Fehler';
			values['error.print_speed.slow_or_fast'] = 'Druckgeschwindigkeit ist falsch (weniger als 10 oder mehr als 1000 mm/s)';
			values['error.first_print_speed.format'] = 'Druckgeschwindigkeit der ersten Schicht - Format Fehler';
			values['error.first_print_speed.slow_or_fast'] = 'Druckgeschwindigkeit der ersten Schicht ist falsch (weniger als 10 oder mehr als 1000 mm/s)';
			values['error.travel_speed.format'] = 'Bewegungsgeschwindigkeit - Format Fehler';
			values['error.travel_speed.slow_or_fast'] = 'Bewegungsgeschwindigkeit ist falsch (weniger als 10 oder mehr als 1000 mm/s)';
			values['error.num_segments.format'] = 'Anzahl der Segmente - Format Fehler';
			values['error.num_segments.slow_or_fast'] = 'Anzahl der Segmente ist falsch (weniger als 2 oder mehr als 100)';
			values['error.init_retract_length.format'] = 'Anfangs-Einzugslänge - Format Fehler';
			values['error.init_retract_length.small_or_big'] = 'Anfangs-Einzugslänge ist falsch (weniger als 0 oder mehr als 20 mm)';
			values['error.end_retract_length.format'] = 'End-Einzugslänge - Format Fehler';
			values['error.end_retract_length.small_or_big'] = 'End-Einzugslänge ist falsch (weniger als 0 oder mehr als 20 mm)';
			values['error.init_retract_speed.format'] = 'Anfangs-Einzugsgeschwindigkeit - Format Fehler';
			values['error.init_retract_speed.slow_or_fast'] = 'Anfangs-Einzugsgeschwindigkeit ist falsch (weniger als 5 oder mehr als 150 mm/s)';
			values['error.end_retract_speed.format'] = 'End-Einzugsgeschwindigkeit - Format Fehler';
			values['error.end_retract_speed.slow_or_fast'] = 'End-Einzugsgeschwindigkeit ist falsch (weniger als 5 oder mehr als 150 mm)';
			values['error.segment_height.format'] = 'Segmenthöhe - Format Fehler';
			values['error.segment_height.small_or_big'] = 'Segmenthöhe ist falsch (weniger als 0.5 oder mehr als 20 mm)';
			values['error.tower_spacing.format'] = 'Abstand zwischen den Türmen - Format Fehler';
			values['error.tower_spacing.too_small'] = 'Abstand zwischen den Türmen ist zu klein';
			values['error.tower_spacing.too_big'] = 'Abstand zwischen den Türmen ist zu groß';
			values['error.z_offset.format'] = 'Z-offset - Format Fehler';
			values['error.z_offset.too_big'] = 'Z-offset-Wert ist falsch (übersteigt die Schichtdicke)';
			values['error.flow.format'] = 'Fluss - Format Fehler';
			values['error.flow.low_or_high'] = 'Wertfehler: Fluss muss zwischen 50 und 150% sein';
			values['error.firmware.not_set'] = 'Format Fehler: Firmware nicht ausgewählt';
			values['error.k_factor.format'] = 'K-Faktor - Format Fehler';
			values['error.k_factor.too_high'] = 'K-Faktor: (Der Wert muss zwischen 0.0 und 2.0 sein)';
			break;
		case 'en':
			values['header.title'] = 'K3D retractions calibrator';
			values['header.description'] = 'You can read a detailed description of the work in <a href="http://k3d.tech/calibrations/retractions/">the article on the main site</a>.';
			values['header.move_exceeds'] = 'If you encounter with error "Move exceeds maximum extrusion", then check <a href="http://k3d.tech/calibrations/retractions/#move-exceeds-maximum-extrusion">here</a>';
			values['header.language'] = 'Language: ';
			
			values['table.header.parameter'] = 'Parameter';
			values['table.header.value'] = 'Value';
			values['table.header.description'] = 'Description';
			
			values['table.bed_size_x.title'] = 'Bed size X';
			values['table.bed_size_x.description'] = '[mm] For cartesian printers - maximum X coordinate<br>For delta-printers - <b>bed diameter</b>';
			values['table.bed_size_y.title'] = 'Bed size Y';
			values['table.bed_size_y.description'] = '[mm] For cartesian printers - maximum Y coordinate<br>For delta-printers - <b>bed diameter</b>';
			values['table.z_offset.title'] = 'Z-offset';
			values['table.z_offset.description'] = '[mm] Offset the entire model vertically. It is necessary to compensate for too thin / thick first layer calibration. Leave zero in general';
			values['table.delta.title'] = 'Origin at the center of the bed';
			values['table.delta.description'] = 'Must be disabled for cartesian printers, enabled for deltas';
			values['table.bed_probe.title'] = 'Bed auto-calibration';
			values['table.bed_probe.description'] = 'Enables bed auto-calibration before printing (G29)? If you don\'t have bed probe, then leave it off';
			values['table.hotend_temp.title'] = 'Hotend temperature';
			values['table.hotend_temp.description'] = '[°C] The temperature to which to heat the hotend before printing';
			values['table.bed_temp.title'] = 'Bed temperature';
			values['table.bed_temp.description'] = '[°C] The temperature to which the bed must be heated before printing. The bed will heat up until parking and auto-calibration.';
			values['table.fan_speed.title'] = 'Fan speed';
			values['table.fan_speed.description'] = '[%] Fan speed in percent. In order to prevent the temperature of the hotend from dropping sharply when the fan is turned on, on the 1st layer it will be turned on by 1/3 of the set value, on the 2nd layer by 2/3, on the 4th layer by the set value';
			values['table.line_width.title'] = 'Line width';
			values['table.line_width.description'] = '[mm] The line width at which the towers will be printed. In general, it is recommended to set equal to the nozzle diameter';
			values['table.first_line_width.title'] = 'First layer line width';
			values['table.first_line_width.description'] = '[mm] The line width at which the raft will be printed under the towers. In general, it is recommended to set 150% of the nozzle diameter';
			values['table.layer_height.title'] = 'Layer height';
			values['table.layer_height.description'] = '[mm] The thickness of the layers of the entire model. In general, 50% of the line width';
			values['table.print_speed.title'] = 'Print speed';
			values['table.print_speed.description'] = '[mm/s] The speed at which towers will be printed';
			values['table.first_print_speed.title'] = 'First layer print speed';
			values['table.first_print_speed.description'] = '[mm/s] The speed at which the raft under the towers will be printed';
			values['table.travel_speed.title'] = 'Travel speed';
			values['table.travel_speed.description'] = '[mm/s] Travel speed between towers';
			values['table.init_retract_length.title'] = 'Initial retraction length';
			values['table.init_retract_length.description'] = '[mm] Retraction length with which the bottom segment will be printed';
			values['table.end_retract_length.title'] = 'Final retraction length';
			values['table.end_retract_length.description'] = '[mm] Retraction length with which the top segment will be printed. Between the lower and upper segments, the retraction length will change in steps for each segment. If you want the length of the retraction to not change, then specify the same value as the initial one';
			values['table.init_retract_speed.title'] = 'Initial retraction speed';
			values['table.init_retract_speed.description'] = '[mm/s] The speed at which retractions will be performed';
			values['table.end_retract_speed.title'] = 'Final retraction speed';
			values['table.end_retract_speed.description'] = '[mm/s] The retraction speed at which the top segment will be printed. Between the bottom and top segment, the retraction speed will change in steps for each segment. If you want the retraction speed to not change, then specify the same value as the initial';
			values['table.num_segments.title'] = 'Number of segments';
			values['table.num_segments.description'] = 'The number of tower segments. During the segment, the length and speed of the retraction remains unchanged. Segments are visually separated to simplify model analysis';
			values['table.segment_height.title'] = 'Segment height';
			values['table.segment_height.description'] = '[mm] The height of one segment of the tower. For example, if the height of the segment is 3mm, and the number of segments is 10, then the height of the entire tower will be 30mm';
			values['table.k_factor.title'] = 'Linear Advance k-factor';
			values['table.k_factor.description'] = 'Enter your value for Linear/Pressure Advance here. If you are not using Linear/Pressure Advance then leave the value at zero';
			values['table.tower_spacing.title'] = 'Distance between towers';
			values['table.tower_spacing.description'] = '[mm] To check retractions, usually about 100 mm is enough. For large printers that often print large models, about half the length of the longer side of the bed is recommended.';
			values['table.firmware.title'] = 'Firmware';
			values['table.firmware.description'] = 'Firmware installed on your printer. If you don\'t know, then it\'s probably Marlin';
			values['table.start_gcode.title'] = 'Start G-Code';
			values['table.start_gcode.description'] = 'The code that is executed before test. Change at your own risk! List of possible placeholders:<br><b>$LA</b> - full command to set the k-factor for LA/PA<br><b>$BEDTEMP</b> - bed temperature<br><b>$HOTTEMP</b> - hotend temperature<br><b>$G29</b> - bed heightmap command<br><b>$FLOW</b> - flow';
			values['table.end_gcode.title'] = 'End G-Code';
			values['table.end_gcode.description'] = 'The code that is executed after the test. Change at your own risk!';
			values['table.hardmode.title'] = 'Hardmode';
			values['table.hardmode.description'] = 'In the normal mode (the parameter is disabled), the tower printing order is generated with optimizations like in a slicer, in the advanced mode - by a non-optimal method. It is recommended to turn it on only when the normal mode shows too optimistic result. Read more in the instructions'
			
			values['generator.generate_and_download'] = 'Generate and download';		
			values['generator.generate_button_loading'] = 'Generator loading...';		
			values['generator.segment'] = ';Segment %d:   %smm @ %smm/s\n';
			values['generator.reset_to_default'] = 'Reset settings';
			
			values['navbar.back'] = ' Back ';
			values['navbar.site'] = 'Site';
			
			values['error.bed_size_x.format'] = 'Bed size Х - format error';
			values['error.bed_size_x.small_or_big'] = 'Bed size X is incorrect (less than 100 or greater than 1000 mm)';
			values['error.bed_size_y.format'] = 'Bed size Y - format error';
			values['error.bed_size_y.small_or_big'] = 'Bed size Y is incorrect (less than 100 or greater than 1000 mm)';
			values['error.hotend_temp.format'] = 'Hotend temperature - format error';
			values['error.hotend_temp.too_low'] = 'Hotend temperature is too low';
			values['error.hotend_temp.too_high'] = 'Hotend temperature is too high';
			values['error.bed_temp.format'] = 'Bed temperature - format error: ';
			values['error.bed_temp.too_high'] = 'Bed temperature is too high';
			values['table.flow.title'] = 'Flow';
			values['table.flow.description'] = '[%] Flow in percents. Needed to compensate for over- or under-extrusion';
			values['error.fan_speed.format'] = 'Fan speed - format error';
			values['error.line_width.format'] = 'Line width - format error';
			values['error.line_width.small_or_big'] = 'Wrong line width (less than 0.1 or greater than 2.0 mm)';
			values['error.first_line_width.format'] = 'First layer line width - format error';
			values['error.first_line_width.small_or_big'] = 'Wrong first line width (less than 0.1 or greater than 2.0 mm)';
			values['error.layer_height.format'] = 'Layer height - format error';
			values['error.layer_height.small_or_big'] = 'Wrong layer height (less than 0.05 mm or greater than 75% from line width)';
			values['error.print_speed.format'] = 'Print speed - format error';
			values['error.print_speed.slow_or_fast'] = 'Wrong print speed (less than 10 or greater than 1000 mm/s)';
			values['error.first_print_speed.format'] = 'First layer print speed - format error';
			values['error.first_print_speed.slow_or_fast'] = 'Wrong first layer print speed (less than 10 or greater than 1000 mm/s)';
			values['error.travel_speed.format'] = 'Travel speed - format error';
			values['error.travel_speed.slow_or_fast'] = 'Wrong travel speed (less than 10 or greater than 1000 mm/s)';
			values['error.num_segments.format'] = 'Number of segments - format error';
			values['error.num_segments.slow_or_fast'] = 'Wrong number of segments (less than 2 or greater than 100)';
			values['error.init_retract_length.format'] = 'Initial retraction length - format error';
			values['error.init_retract_length.small_or_big'] = 'Wrong initial retraction length (less than 0 or greater than 20 mm)';
			values['error.end_retract_length.format'] = 'Final retraction length - format error';
			values['error.end_retract_length.small_or_big'] = 'Wrong final retraction length (less than 0 or greater than 20 mm)';
			values['error.init_retract_speed.format'] = 'Initial retraction speed - format error';
			values['error.init_retract_speed.slow_or_fast'] = 'Wrong initial retraction speed (less than 5 or greater than 150 mm/s)';
			values['error.end_retract_speed.format'] = 'Final retraction speed - format error';
			values['error.end_retract_speed.slow_or_fast'] = 'Wrong final retraction speed (less than 5 or greater than 150 mm/s)';
			values['error.segment_height.format'] = 'Segment height - format error';
			values['error.segment_height.small_or_big'] = 'Wrong segment height (less than 0.5 or greater than 20 mm)';
			values['error.tower_spacing.format'] = 'Distance between towers - format error';
			values['error.tower_spacing.too_small'] = 'Distance between towers is too low';
			values['error.tower_spacing.too_big'] = 'Distance between towers is too high';
			values['error.z_offset.format'] = 'Z-offset - format error';
			values['error.z_offset.too_big'] = 'Offset value is wrong (exceeds the layer thickness in absolute value)';
			values['error.flow.format'] = 'Flow - format error';
			values['error.flow.low_or_high'] = 'Value error: flow should be from 50 to 150%';
			values['error.firmware.not_set'] = 'Format error: firmware not set';
			values['error.k_factor.format'] = 'K-factor - format error';
			values['error.k_factor.too_high'] = 'Wrong K-factor value (should be from 0.0 to 2.0)';
			break;
		case 'ru':
			values['header.title'] = 'K3D калибровщик откатов';
			values['header.description'] = 'Подробное описание работы вы можете прочитать в <a href="http://k3d.tech/calibrations/retractions/">статье на основном сайте.</a>';
			values['header.move_exceeds'] = 'Если сталкиваетесь с ошибкой "Move exceeds maximum extrusion", то вам <a href="http://k3d.tech/calibrations/retractions/#move-exceeds-maximum-extrusion">сюда</a>';
			values['header.language'] = 'Язык: ';
			
			values['table.header.parameter'] = 'Параметр';
			values['table.header.value'] = 'Значение';
			values['table.header.description'] = 'Описание';
			
			values['table.bed_size_x.title'] = 'Размер стола по X';
			values['table.bed_size_x.description'] = '[мм] Для декартовых принтеров - максимальная координата по оси X<br>Для дельта-принтеров - <b>диаметр стола</b>';
			values['table.bed_size_y.title'] = 'Размер стола по Y';
			values['table.bed_size_y.description'] = '[мм] Для декартовых принтеров - максимальная координата по оси Y<br>Для дельта-принтеров - <b>диаметр стола</b>';
			values['table.z_offset.title'] = 'Z-offset';
			values['table.z_offset.description'] = '[мм] Смещение всей модели по вертикали. Нужно чтобы компенсировать слишком тонкую/толстую калибровку первого слоя. В общем случае оставьте ноль';
			values['table.delta.title'] = 'Начало координат в центре стола';
			values['table.delta.description'] = 'Для декартовых принтеров должно быть выключено, для дельт включено';
			values['table.bed_probe.title'] = 'Автокалибровка стола';
			values['table.bed_probe.description'] = 'Надо ли делать автокалибровку стола перед печатью (G29)? Если у вас нет датчика автокалибровки, то оставляйте выключенным';
			values['table.hotend_temp.title'] = 'Температура хотэнда';
			values['table.hotend_temp.description'] = '[°C] Температура, до которой нагреть хотэнд перед печатью';
			values['table.bed_temp.title'] = 'Температура стола';
			values['table.bed_temp.description'] = '[°C] Температура, до которой нагреть стол перед печатью. Стол будет нагрет до выполнения парковки и автокалибровки стола';
			values['table.flow.title'] = 'Поток';
			values['table.flow.description'] = '[%] Поток в процентах. Нужен для компенсации пере- или недоэкструзии';
			values['table.fan_speed.title'] = 'Скорость вентилятора';
			values['table.fan_speed.description'] = '[%] Обороты вентилятора в процентах. Для того, чтобы температура хотэнда резко не упала при включении вентилятора, на 1 слое он будет включен на 1/3 от заданного значения, на 2 слое на 2/3, на 4 слое на заданное значение';
			values['table.line_width.title'] = 'Ширина линии';
			values['table.line_width.description'] = '[мм] Ширина линий, с которой будут напечатаны башенки. В общем случае рекомендуется выставить равной диаметру сопла';
			values['table.first_line_width.title'] = 'Ширина линии первого слоя';
			values['table.first_line_width.description'] = '[мм] Ширина линий, с которой будет напечатана подложка под башенками. В общем случае рекомендуется выставить 150% от диаметра сопла';
			values['table.layer_height.title'] = 'Толщина слоя';
			values['table.layer_height.description'] = '[мм] Толщина слоёв всей модели. В общем случае 50% от ширины линии';
			values['table.print_speed.title'] = 'Скорость печати';
			values['table.print_speed.description'] = '[мм/с] Скорость, с которой будут напечатаны башенки';
			values['table.first_print_speed.title'] = 'Скорость печати первого слоя';
			values['table.first_print_speed.description'] = '[мм/с] Скорость, с которой будет напечатана подложка под башенки';
			values['table.travel_speed.title'] = 'Скорость перемещений';
			values['table.travel_speed.description'] = '[мм/с] Скорость перемещений между башенками';
			values['table.init_retract_length.title'] = 'Начальная длина отката';
			values['table.init_retract_length.description'] = '[мм] Длина отката, с которой будет напечатан нижний сегмент';
			values['table.end_retract_length.title'] = 'Конечная длина отката';
			values['table.end_retract_length.description'] = '[мм] Длина отката, с которой будет напечатан верхний сегмент. Между нижним и верхним сегментом длина отката будет изменяться ступенчато за каждый сегмент. Если хотите, чтобы длина отката не менялась, то укажите такое же значение, как начальное';
			values['table.init_retract_speed.title'] = 'Начальная скорость отката';
			values['table.init_retract_speed.description'] = '[мм/с] Скорость, с которой будут выполняться откаты';
			values['table.end_retract_speed.title'] = 'Конечная скорость отката';
			values['table.end_retract_speed.description'] = '[мм/с] Скорость отката, с которой будет напечатан верхний сегмент. Между нижним и верхним сегментом скорость отката будет изменяться ступенчато за каждый сегмент. Если хотите, чтобы скорость отката не менялась, то укажите такое же значение, как начальное';
			values['table.num_segments.title'] = 'Количество сегментов';
			values['table.num_segments.description'] = 'Количество сегментов башенки. В течение сегмента длина и скорость отката остаются неизменными. Сегменты визуально разделены для упрощения анализа модели';
			values['table.segment_height.title'] = 'Высота сегмента';
			values['table.segment_height.description'] = '[мм] Высота одного сегмента башенки. К примеру, если высота сегмента 3мм, а количество сегментов 10, то высота всей башенки будет 30мм';
			values['table.k_factor.title'] = 'k-фактор Linear Advance';
			values['table.k_factor.description'] = 'Введите сюда ваше значение для Linear/Pressure Advance. Если вы не пользуетесь Linear/Pressure Advance, то оставьте значение нулевым';
			values['table.tower_spacing.title'] = 'Расстояние между башенками';
			values['table.tower_spacing.description'] = '[мм] Для проверки откатов, обычно, хватает около 100 мм. Для крупногабаритных принтеров, которые часто печатают большие модели, рекомендуется около половины длины большей стороны стола';
			values['table.firmware.title'] = 'Прошивка';
			values['table.firmware.description'] = 'Прошивка, установленная на вашем принтере. Если не знаете, то, скорее всего, Marlin';
			values['table.start_gcode.title'] = 'Начальный G-код';
			values['table.start_gcode.description'] = 'Код, выполняемый перед печатью теста. Менять на свой страх и риск! Список возможных плейсхолдеров:<br><b>$LA</b> - полная команда на установку к-фактора LA/PA<br><b>$BEDTEMP</b> - температура стола<br><b>$HOTTEMP</b> - температура хотэнда<br><b>$G29</b> - команда на снятие карты высот стола<br><b>$FLOW</b> - поток';
			values['table.end_gcode.title'] = 'Конечный G-код';
			values['table.end_gcode.description'] = 'Код, выполняемый после печати теста. Менять на свой страх и риск!';
			values['table.hardmode.title'] = 'Усложненный режим';
			values['table.hardmode.description'] = 'В обычном режиме (параметр выключен) порядок печати башен генерируется с оптимизациями как в слайсере, в усложненном - неоптимальным методом. Рекомендуется включать только тогда, когда обычный режим показывает слишком оптимистичный результат. Подробнее в инструкции';
			
			values['generator.generate_and_download'] = 'Генерировать и скачать';		
			values['generator.generate_button_loading'] = 'Генератор загружается...';		
			values['generator.segment'] = ';Сегмент %d:   %sмм @ %sмм/с\n';
			values['generator.reset_to_default'] = 'Сбросить настройки';
			
			values['navbar.back'] = ' Назад ';
			values['navbar.site'] = 'Сайт';
			
			values['error.bed_size_x.format'] = 'Размер оси Х - ошибка формата';
			values['error.bed_size_x.small_or_big'] = 'Размер стола по X указан неверно (меньше 100 или больше 1000 мм)';
			values['error.bed_size_y.format'] = 'Размер оси Y - ошибка формата';
			values['error.bed_size_y.small_or_big'] = 'Размер стола по Y указан неверно (меньше 100 или больше 1000 мм)';
			values['error.hotend_temp.format'] = 'Температура хотэнда - ошибка формата';
			values['error.hotend_temp.too_low'] = 'Температура хотэнда слишком низкая';
			values['error.hotend_temp.too_high'] = 'Температура хотэнда слишком высокая';
			values['error.bed_temp.format'] = 'Температура стола - ошибка формата: ';
			values['error.bed_temp.too_high'] = 'Температура стола слишком высокая';
			values['error.fan_speed.format'] = 'Скорость вентилятора - ошибка формата';
			values['error.line_width.format'] = 'Ширина линии - ошибка формата';
			values['error.line_width.small_or_big'] = 'Неправильная ширина линии (меньше 0.1 или больше 2.0 мм)';
			values['error.first_line_width.format'] = 'Ширина линии первого слоя - ошибка формата';
			values['error.first_line_width.small_or_big'] = 'Неправильная ширина линии первого слоя (меньше 0.1 или больше 2.0 мм)';
			values['error.layer_height.format'] = 'Высота слоя - ошибка формата';
			values['error.layer_height.small_or_big'] = 'Толщина слоя неправильная (меньше 0.05 мм или больше 75% от ширины линии)';
			values['error.print_speed.format'] = 'Скорость печати - ошибка формата';
			values['error.print_speed.slow_or_fast'] = 'Скорость печати неправильная (меньше 10 или больше 1000 мм/с)';
			values['error.first_print_speed.format'] = 'Скорость печати первого слоя - ошибка формата';
			values['error.first_print_speed.slow_or_fast'] = 'Скорость печати первого слоя неправильная (меньше 10 или больше 1000 мм/с)';
			values['error.travel_speed.format'] = 'Скорость перемещений - ошибка формата';
			values['error.travel_speed.slow_or_fast'] = 'Скорость перемещений неправильная (меньше 10 или больше 1000 мм/с)';
			values['error.num_segments.format'] = 'Количество сегментов - ошибка формата';
			values['error.num_segments.slow_or_fast'] = 'Количество сегментов неправильное (меньше 2 или больше 100)';
			values['error.init_retract_length.format'] = 'Начальная длина отката - ошибка формата';
			values['error.init_retract_length.small_or_big'] = 'Начальная длина отката неправильная (меньше 0 или больше 20 мм)';
			values['error.end_retract_length.format'] = 'Конечная длина отката - ошибка формата';
			values['error.end_retract_length.small_or_big'] = 'Конечная длина отката неправильная (меньше 0 или больше 20 мм)';
			values['error.init_retract_speed.format'] = 'Начальная скорость отката - ошибка формата';
			values['error.init_retract_speed.slow_or_fast'] = 'Начальная скорость отката неправильная (меньше 5 или больше 150 мм/с)';
			values['error.end_retract_speed.format'] = 'Конечная скорость отката - ошибка формата';
			values['error.end_retract_speed.slow_or_fast'] = 'Конечная скорость отката неправильная (меньше 5 или больше 150 мм)';
			values['error.segment_height.format'] = 'Высота сегмента - ошибка формата';
			values['error.segment_height.small_or_big'] = 'Высота сегмента неправильная (меньше 0.5 или больше 20 мм)';
			values['error.tower_spacing.format'] = 'Расстояние между башенками - ошибка формата';
			values['error.tower_spacing.too_small'] = 'Расстояние между башенками слишком мало';
			values['error.tower_spacing.too_big'] = 'Расстояние между башенками слишком велико';
			values['error.z_offset.format'] = 'Z-offset - ошибка формата';
			values['error.z_offset.too_big'] = 'Значение оффсета неправильно (превышает толщину слоя по модулю)';
			values['error.flow.format'] = 'Поток - ошибка формата';
			values['error.flow.low_or_high'] = 'Ошибка значения: поток должен быть от 50 до 150%';
			values['error.firmware.not_set'] = 'Ошибка формата: не выбрана прошивка';
			values['error.k_factor.format'] = 'K-фактор - ошибка формата';
			values['error.k_factor.too_high'] = 'Неверное значение K-фактора (должно быть от 0.0 до 2.0)';
			break;
	}
	
	document.title = window.lang.getString('header.title');
	var el = document.getElementsByClassName('lang');
	for (var i = 0; i < el.length; i++) {
		var item = el[i];
		item.innerHTML = window.lang.getString(item.id);
	}
	document.getElementsByClassName('generate-button')[0].innerHTML = window.lang.getString('generator.generate_and_download');
	document.getElementsByClassName('reset-button')[0].innerHTML = window.lang.getString('generator.reset_to_default');
	document.getElementsByClassName('navbar-direction')[0].innerHTML = window.lang.getString('navbar.back');
	document.getElementById('generateButtonLoading').innerHTML = window.lang.getString('generator.generate_button_loading');
}

function setSegmentsPreview(segments) {
	savedSegmentsInfo = segments;
	setSegmentsPreviewVisible(true);
}

function setSegmentsPreviewVisible(visible) {
	if (savedSegmentsInfo == null || savedSegmentsInfo == undefined) {
		visible = false;
	}
	if (visible) {
		document.getElementById('table.' + segmentKeys[0] + '.description').rowSpan = segmentKeys.length;
		document.getElementById('table.' + segmentKeys[0] + '.description').innerHTML = '<span>' + savedSegmentsInfo.replaceAll('\n', '<br>') + '</span>';
		
		for (var i = 1; i < segmentKeys.length; i++) {
			document.getElementById('table.' + segmentKeys[i] + '.description').style.display = 'none';
		}
	} else {
		document.getElementById('table.' + segmentKeys[0] + '.description').rowSpan = 1;
		for (var i = 0; i < segmentKeys.length; i++) {
			var id = 'table.' + segmentKeys[i] + '.description';
			document.getElementById(id).style.display = '';
			document.getElementById(id).innerHTML = window.lang.getString(id);
		}
	}
}

function reset() {
	for (var elementId of formFields) {
        localStorage.removeItem(elementId);
    }
	
	window.location.reload(false);
}

function init() {
	initForm();
	
	const urlParams = new URLSearchParams(window.location.search);
	var lang = urlParams.get('lang');
	if (lang == undefined) {
		lang = 'ru';
	}
	
	window.lang = {
		values: {},
		getString: function(key) {
			var ret = window.lang.values[key];
			if (key == 'header.title') {
				return ret + ' ' + calibrator_version;
			}
			return ret;
		}
	};
	initLang(lang);
	
	setTimeout(function() {
		if (checkGo != undefined && window.lang != undefined) {
			checkGo();
		} else {
			setTimeout(this, 100);
		}
	}, 100);
}