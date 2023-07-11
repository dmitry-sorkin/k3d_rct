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
    "kFactor",
    "towerSpacing",
];

var saveForm = function () {
    for (var elementId of formFields) {
        var element = document.getElementById(elementId);
        if (element) {
            var saveValue = element.value;
            if (elementId == 'delta' || elementId == 'bedProbe') {
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
            if (elementId == 'delta' || elementId == 'bedProbe') {
                if (loadValue == 'true') {
                    element.checked = true;
                } else {
                    element.checked = false;
                }
                
            } else {
                if (loadValue != null) {
                    element.value = loadValue;
                }
            }
            
        }
    }
}

function initForm() {
    for (var elementId of formFields) {
        var element = document.getElementById(elementId);
        element.onchange = saveForm;
    }
    loadForm();
}
