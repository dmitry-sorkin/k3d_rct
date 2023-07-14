package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"syscall/js"
)

const filamentDiameter = 1.75

var (
	bedX, bedY, lineWidth, firstLayerLineWidth, printSpeed, travelSpeed, layerHeight, initRetractLength, retractLength, retractLengthDelta, currentE, firstLayerPrintSpeed, segmentHeight, towerSpacing, towerWidth, zOffset, initRetractSpeed, retractSpeed, currentSpeed, retractSpeedDelta, kFactor float64
	hotendTemperature, bedTemperature, numSegments, cooling, flow, firmware                                                                                                                                                                                                                            int
	currentCoordinates                                                                                                                                                                                                                                                                                 Point
	bedProbe, retracted, delta                                                                                                                                                                                                                                                                         bool
	startGcode, endGcode                                                                                                                                                                                                                                                                               string
)

type Point struct {
	X float64
	Y float64
	Z float64
}

func main() {
	c := make(chan struct{})
	registerFunctions()
	<-c
}

func registerFunctions() {
	js.Global().Set("generate", js.FuncOf(generate))
	js.Global().Set("checkGo", js.FuncOf(checkJs))
}

func setErrorDescription(doc js.Value, lang js.Value, key string, curErr string, hasErr bool) {
	if hasErr {
		doc.Call("getElementById", key).Set("innerHTML", lang.Call("getString", key).String() + "<br><span class=\"inline-error\">" + curErr + "</span>")
	} else {
		doc.Call("getElementById", key).Set("innerHTML", lang.Call("getString", key).String())
	}
}

func check(showErrorBox bool) bool {
	errorString := ""
	doc := js.Global().Get("document")
	lang := js.Global().Get("lang")
	doc.Call("getElementById", "resultContainer").Set("innerHTML", "")

	// Fill variables with data from web page
	curErr := ""
	hasErr := false
	
	docBedX, err := parseInputToFloat(doc.Call("getElementById", "bedX").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.bed_size_x.format").String(), true
	} else if docBedX < 100 || docBedX > 1000 {
		curErr, hasErr = lang.Call("getString", "error.bed_size_x.small_or_big").String(), true
	} else {
		bedX = docBedX
	}
	
	setErrorDescription(doc, lang, "table.bed_size_x.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docBedY, err := parseInputToFloat(doc.Call("getElementById", "bedY").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.bed_size_y.format").String(), true
	} else if docBedY < 100 || docBedY > 1000 {
		curErr, hasErr = lang.Call("getString", "error.bed_size_y.small_or_big").String(), true
	} else {
		bedY = docBedY
	}
	setErrorDescription(doc, lang, "table.bed_size_y.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	delta = doc.Call("getElementById", "delta").Get("checked").Bool()

	bedProbe = doc.Call("getElementById", "bedProbe").Get("checked").Bool()

	docHotTemp, err := parseInputToInt(doc.Call("getElementById", "hotendTemperature").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.hotend_temp.format").String(), true
	} else if docHotTemp < 150 {
		curErr, hasErr = lang.Call("getString", "error.hotend_temp.too_low").String(), true
	} else if docHotTemp > 350 {
		curErr, hasErr = lang.Call("getString", "error.hotend_temp.too_high").String(), true
	} else {
		hotendTemperature = docHotTemp
	}
	setErrorDescription(doc, lang, "table.hotend_temp.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docBedTemp, err := parseInputToInt(doc.Call("getElementById", "bedTemperature").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.bed_temp.format").String() + err.Error(), true
	} else if docBedTemp > 150 {
		curErr, hasErr = lang.Call("getString", "error.bed_temp.too_high").String(), true
	} else {
		bedTemperature = docBedTemp
	}
	setErrorDescription(doc, lang, "table.bed_temp.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docCooling, err := parseInputToInt(doc.Call("getElementById", "cooling").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.fan_speed.format").String(), true
	} else {
		docCooling = int(float64(docCooling) * 2.55)
		if docCooling < 0 {
			docCooling = 0
		} else if docCooling > 255 {
			docCooling = 255
		}
		cooling = docCooling
	}
	setErrorDescription(doc, lang, "table.fan_speed.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docLineWidth, err := parseInputToFloat(doc.Call("getElementById", "lineWidth").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.line_width.format").String(), true
	} else if docLineWidth < 0.1 || docLineWidth > 2.0 {
		curErr, hasErr = lang.Call("getString", "error.line_width.small_or_big").String(), true
	} else {
		lineWidth = docLineWidth
	}
	setErrorDescription(doc, lang, "table.line_width.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docFirstLineWidth, err := parseInputToFloat(doc.Call("getElementById", "firstLayerLineWidth").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.first_line_width.format").String(), true
	} else if docFirstLineWidth < 0.1 || docFirstLineWidth > 2.0 {
		curErr, hasErr = lang.Call("getString", "error.first_line_width.small_or_big").String(), true
	} else {
		firstLayerLineWidth = docFirstLineWidth
	}
	setErrorDescription(doc, lang, "table.first_line_width.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docLayerHeight, err := parseInputToFloat(doc.Call("getElementById", "layerHeight").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.layer_height.format").String(), true
	} else if docLayerHeight < 0.05 || docLayerHeight > lineWidth*0.75 {
		curErr, hasErr = lang.Call("getString", "error.layer_height.small_or_big").String(), true
	} else {
		layerHeight = docLayerHeight
	}
	setErrorDescription(doc, lang, "table.layer_height.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docPrintSpeed, err := parseInputToFloat(doc.Call("getElementById", "printSpeed").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.print_speed.format").String(), true
	} else if docPrintSpeed < 10 || docPrintSpeed > 1000 {
		curErr, hasErr = lang.Call("getString", "error.print_speed.slow_or_fast").String(), true
	} else {
		printSpeed = docPrintSpeed
	}
	setErrorDescription(doc, lang, "table.print_speed.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docFirstPrintSpeed, err := parseInputToFloat(doc.Call("getElementById", "firstLayerPrintSpeed").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.first_print_speed.format").String(), true
	} else if docFirstPrintSpeed < 10 || docFirstPrintSpeed > 1000 {
		curErr, hasErr = lang.Call("getString", "error.first_print_speed.slow_or_fast").String(), true
	} else {
		firstLayerPrintSpeed = docFirstPrintSpeed
	}
	setErrorDescription(doc, lang, "table.first_print_speed.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docTravelSpeed, err := parseInputToFloat(doc.Call("getElementById", "travelSpeed").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.travel_speed.format").String(), true
	} else if docTravelSpeed < 10 || docTravelSpeed > 1000 {
		curErr, hasErr = lang.Call("getString", "error.travel_speed.slow_or_fast").String(), true
	} else {
		travelSpeed = docTravelSpeed
	}
	setErrorDescription(doc, lang, "table.travel_speed.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docNumSegments, err := parseInputToInt(doc.Call("getElementById", "numSegments").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.num_segments.format").String(), true
	} else if docNumSegments < 2 || docNumSegments > 100 {
		curErr, hasErr = lang.Call("getString", "error.num_segments.slow_or_fast").String(), true
	} else {
		numSegments = docNumSegments
	}
	setErrorDescription(doc, lang, "table.num_segments.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docInitRetractLength, err := parseInputToFloat(doc.Call("getElementById", "initRetractLength").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.init_retract_length.format").String(), true
	} else if docInitRetractLength < 0 || docInitRetractLength > 20 {
		curErr, hasErr = lang.Call("getString", "error.init_retract_length.small_or_big").String(), true
	} else {
		retractLength = docInitRetractLength
		initRetractLength = docInitRetractLength
	}
	setErrorDescription(doc, lang, "table.init_retract_length.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docEndRetractLength, err := parseInputToFloat(doc.Call("getElementById", "endRetractLength").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.end_retract_length.format").String(), true
	} else if docEndRetractLength < 0 || docEndRetractLength > 20 {
		curErr, hasErr = lang.Call("getString", "error.end_retract_length.small_or_big").String(), true
	} else {
		retractLengthDelta = (docInitRetractLength - docEndRetractLength) / float64(numSegments-1)
	}
	setErrorDescription(doc, lang, "table.end_retract_length.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docRetractSpeed, err := parseInputToFloat(doc.Call("getElementById", "initRetractSpeed").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.init_retract_speed.format").String(), true
	} else if docRetractSpeed < 5 || docRetractSpeed > 150 {
		curErr, hasErr = lang.Call("getString", "error.init_retract_speed.slow_or_fast").String(), true
	} else {
		retractSpeed = docRetractSpeed
		initRetractSpeed = docRetractSpeed
	}
	setErrorDescription(doc, lang, "table.init_retract_speed.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docEndRetractSpeed, err := parseInputToFloat(doc.Call("getElementById", "endRetractSpeed").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.end_retract_speed.format").String(), true
	} else if docEndRetractSpeed < 5 || docEndRetractSpeed > 150 {
		curErr, hasErr = lang.Call("getString", "error.end_retract_speed.slow_or_fast").String(), true
	} else {
		retractSpeedDelta = (docRetractSpeed - docEndRetractSpeed) / float64(numSegments-1)
	}
	setErrorDescription(doc, lang, "table.end_retract_speed.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docSegmentHeight, err := parseInputToFloat(doc.Call("getElementById", "segmentHeight").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.segment_height.format").String(), true
	} else if docSegmentHeight < 0.5 || docSegmentHeight > 20 {
		curErr, hasErr = lang.Call("getString", "error.segment_height.small_or_big").String(), true
	} else {
		segmentHeight = docSegmentHeight
	}
	setErrorDescription(doc, lang, "table.segment_height.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docTowerSpacing, err := parseInputToFloat(doc.Call("getElementById", "towerSpacing").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.tower_spacing.format").String(), true
	} else if docTowerSpacing < 40 {
		curErr, hasErr = lang.Call("getString", "error.tower_spacing.too_small").String(), true
	} else if docTowerSpacing > bedX-40.0 {
		curErr, hasErr = lang.Call("getString", "error.tower_spacing.too_big").String(), true
	} else {
		towerSpacing = docTowerSpacing
	}
	setErrorDescription(doc, lang, "table.tower_spacing.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}

	docZOffset, err := parseInputToFloat(doc.Call("getElementById", "zOffset").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.z_offset.format").String(), true
	} else if docZOffset < -layerHeight || docZOffset > layerHeight {
		curErr, hasErr = lang.Call("getString", "error.z_offset.too_big").String(), true
	} else {
		zOffset = docZOffset
	}
	setErrorDescription(doc, lang, "table.z_offset.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}
	
	docFlow, err := parseInputToInt(doc.Call("getElementById", "flow").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.flow.format").String(), true
	} else if docFlow < 50 || docFlow > 150 {
		curErr, hasErr = lang.Call("getString", "error.flow.low_or_high").String(), true
	} else {
		flow = docFlow
	}
	setErrorDescription(doc, lang, "table.flow.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}
	
	docKFactor, err := parseInputToFloat(doc.Call("getElementById", "kFactor2").Get("value").String())
	if err != nil {
		curErr, hasErr = lang.Call("getString", "error.k_factor.format").String(), true
	} else if docKFactor < 0.0 || docKFactor > 2.0 {
		curErr, hasErr = lang.Call("getString", "error.k_factor.too_high").String(), true
	} else {
		kFactor = docKFactor
	}
	setErrorDescription(doc, lang, "table.k_factor.description", curErr, hasErr)
	if hasErr {
		errorString = errorString + curErr + "\n"
		hasErr = false
	}
	
	docMarlin := doc.Call("getElementById", "firmwareMarlin").Get("checked").Bool()
	docKlipper := doc.Call("getElementById", "firmwareKlipper").Get("checked").Bool()
	docRRF := doc.Call("getElementById", "firmwareRRF").Get("checked").Bool()
	if docMarlin {
		firmware = 0
	} else if docKlipper {
		firmware = 1
	} else if docRRF {
		firmware = 2
	} else {
		errorString = errorString + lang.Call("getString", "error.firmware.not_set").String() + "\n"
	}
	
	startGcode = doc.Call("getElementById", "startGcode").Get("innerHTML").String()
	endGcode = doc.Call("getElementById", "endGcode").Get("innerHTML").String()
	
	if !showErrorBox {
		return true
	}

	// end check of parameters
	if errorString == "" {
		println("OK")
		return true
	} else {
		println(errorString)
		js.Global().Call("showError", errorString)
		return false
	}
}

func checkJs(this js.Value, i []js.Value) interface{} {
	check(false)
	return js.ValueOf(nil)
}

func generate(this js.Value, i []js.Value) interface{} {
	// check and initialize variables
	if check(true) {
		lang := js.Global().Get("lang")
		segmentStr := lang.Call("getString", "generator.segment").String()
		
		// generate calibration parameters
		caliParams := ""
		for i := numSegments - 1; i >= 0; i-- {
			caliParams = caliParams + fmt.Sprintf(segmentStr,
				i+1,
				fmt.Sprint(roundFloat(initRetractLength-retractLengthDelta*float64(i), 2)),
				fmt.Sprint(roundFloat(initRetractSpeed-retractSpeedDelta*float64(i), 2)))
		}

		gcode := make([]string, 0, 1)
		// gcode initialization
		gcode = append(gcode, "; generated by K3D Retraction calibration towers generator ",
			js.Global().Get("calibrator_version").String(),
			"\n",
			"; Written by Dmitry Sorkin @ http://k3d.tech/\n",
			"; and Kekht\n",
			fmt.Sprintf(";Bedsize: %f:%f\n", bedX, bedY),
			fmt.Sprintf(";Temp: %d/%d\n", hotendTemperature, bedTemperature),
			fmt.Sprintf(";Width: %f-%f\n", lineWidth, firstLayerLineWidth),
			fmt.Sprintf(";Layer height: %f\n", layerHeight),
			fmt.Sprintf(";Retract length: %f, -%f/segment\n", retractLength, retractLengthDelta),
			fmt.Sprintf(";Segments: %dx%f mm\n", numSegments, segmentHeight),
			caliParams)
			
		var g29 string
		if bedProbe {
			g29 = "G29"
		} else {
			g29 = ""
		}
		replacer := strings.NewReplacer("$LA", generateLACommand(kFactor), "$BEDTEMP", strconv.Itoa(bedTemperature), "$HOTTEMP", strconv.Itoa(hotendTemperature), "$G29", g29, "$FLOW", strconv.Itoa(flow))
		gcode = append(gcode, replacer.Replace(startGcode), "\n")
		
		gcode = append(gcode, "M82\n", fmt.Sprintf("M106 S%d\n", int(cooling/3)))

		// generate first layer
		var bedCenter, leftTowerCenter, rightTowerCenter Point
		if delta {
			bedCenter.X, bedCenter.Y, bedCenter.Z = 0, 0, layerHeight
		} else {
			bedCenter.X, bedCenter.Y, bedCenter.Z = bedX/2, bedY/2, layerHeight
		}
		leftTowerCenter = bedCenter
		leftTowerCenter.X = leftTowerCenter.X - towerSpacing/2
		rightTowerCenter = bedCenter
		rightTowerCenter.X = bedCenter.X + towerSpacing/2
		currentE = 0
		currentSpeed = firstLayerPrintSpeed
		currentCoordinates.X, currentCoordinates.Y, currentCoordinates.Z = 0, 0, 0

		// purge nozzle
		var purgeStart Point
		purgeStart.X, purgeStart.Y, purgeStart.Z = leftTowerCenter.X-15.0, leftTowerCenter.Y-25.0, layerHeight
		purgeTwo := purgeStart
		purgeTwo.X = rightTowerCenter.X + 15.0
		purgeThree := purgeTwo
		purgeThree.Y += firstLayerLineWidth
		purgeEnd := purgeThree
		purgeEnd.X = purgeStart.X

		// move Z to first layer coordinates
		gcode = append(gcode, fmt.Sprintf("G1 Z%s F450\n", fmt.Sprint(roundFloat(layerHeight+zOffset, 2))))
		currentSpeed = 450/60

		// make printer think, that he is on layerHeight
		gcode = append(gcode, fmt.Sprintf("G92 Z%s\n", fmt.Sprint(roundFloat(layerHeight, 2))))
		currentCoordinates.Z = layerHeight

		// move to start of purge
		gcode = append(gcode, generateMove(currentCoordinates, purgeStart, 0.0)...)

		// add purge to gcode
		gcode = append(gcode, generateMove(currentCoordinates, purgeTwo, firstLayerLineWidth)...)
		gcode = append(gcode, generateMove(currentCoordinates, purgeThree, firstLayerLineWidth)...)
		gcode = append(gcode, generateMove(currentCoordinates, purgeEnd, firstLayerLineWidth)...)

		// generate raft trajectory for left tower
		trajectory := generateZigZagTrajectory(leftTowerCenter, firstLayerLineWidth)

		// move to start of left tower raft
		gcode = append(gcode, generateMove(currentCoordinates, trajectory[0], 0.0)...)

		// print left tower raft
		for i := 1; i < len(trajectory); i++ {
			gcode = append(gcode, generateMove(currentCoordinates, trajectory[i], firstLayerLineWidth)...)
		}

		// generate raft trajectory for right tower
		for i := 0; i < len(trajectory); i++ {
			trajectory[i].X = trajectory[i].X + towerSpacing
		}

		// move to start of right tower raft
		gcode = append(gcode, generateMove(currentCoordinates, trajectory[0], 0.0)...)

		// print right tower raft
		for i := 1; i < len(trajectory); i++ {
			gcode = append(gcode, generateMove(currentCoordinates, trajectory[i], firstLayerLineWidth)...)
		}

		// generate towers
		layersPerSegment := int(segmentHeight / layerHeight)
		for i := 1; i < numSegments*layersPerSegment; i++ {
			// set new layer coordinates
			currentCoordinates.Z += layerHeight

			// add layer start comment
			gcode = append(gcode, fmt.Sprintf(";layer #%s\n", fmt.Sprint(roundFloat(currentCoordinates.Z/layerHeight, 0))))

			// change fan speed
			if i == 1 {
				gcode = append(gcode, fmt.Sprintf("M106 S%d\n", int(cooling*2/3)))
			} else if i == 2 {
				gcode = append(gcode, fmt.Sprintf("M106 S%d\n", cooling))
			}

			// modify print settings if switching segments
			if i%layersPerSegment == 0 {
				retractLength = retractLength - retractLengthDelta
				if retractLength < 0.1 {
					retractLength = 0.1
				}
				retractSpeed = retractSpeed - retractSpeedDelta
				if retractSpeed < 5 {
					retractSpeed = 5
				}
				towerWidth = 15.0 + lineWidth/2
			} else {
				towerWidth = 15.0
			}

			// interchange tower centers on odd layers
			firstTowerCenter := rightTowerCenter
			secondTowerCenter := leftTowerCenter
			if i%2 == 0 {
				firstTowerCenter = leftTowerCenter
				secondTowerCenter = rightTowerCenter
			}

			// generate first tower trajectory
			trajectory = generateSquareTrajectory(firstTowerCenter, towerWidth-2.3*lineWidth)
			trajectory = append(trajectory, generateSquareTrajectory(firstTowerCenter, towerWidth-0.5*lineWidth)...)

			// if first tower is right tower, that rotate it CCW
			if firstTowerCenter == rightTowerCenter {
				trajectory = rotateSquareTrajectoryCW(trajectory)
			}

			// move to start of first tower
			gcode = append(gcode, generateMove(currentCoordinates, trajectory[0], 0.0)...)

			// move to new layer
			gcode = append(gcode, fmt.Sprintf("G1 Z%s F300\n", fmt.Sprint(roundFloat(currentCoordinates.Z, 2))))
			currentSpeed = 300/60

			// print first tower
			for i := 1; i < len(trajectory); i++ {
				gcode = append(gcode, generateMove(currentCoordinates, trajectory[i], lineWidth)...)
			}

			// generate second tower trajectory
			trajectory = generateSquareTrajectory(secondTowerCenter, towerWidth-2.3*lineWidth)
			trajectory = append(trajectory, generateSquareTrajectory(secondTowerCenter, towerWidth-0.5*lineWidth)...)

			// if second tower is right tower, rotate second tower trajectory CCW
			if secondTowerCenter == rightTowerCenter {
				trajectory = rotateSquareTrajectoryCW(trajectory)
			}

			// move to start of second tower
			gcode = append(gcode, generateMove(currentCoordinates, trajectory[0], 0.0)...)

			// print second tower
			for i := 1; i < len(trajectory); i++ {
				gcode = append(gcode, generateMove(currentCoordinates, trajectory[i], lineWidth)...)
			}
		}

		// end gcode
		gcode = append(gcode, ";end gcode\n", replacer.Replace(endGcode))


		outputGCode := ""
		for i := 0; i < len(gcode); i++ {
			outputGCode = outputGCode + gcode[i]
		}

		// write calibration parameters to resultContainer
		js.Global().Call("showError", caliParams)

		// save file
		fileName := fmt.Sprintf("K3D_RCT_H%d-B%d_%s-%smm_%s-%smms.gcode",
			hotendTemperature,
			bedTemperature,
			fmt.Sprint(roundFloat(initRetractLength, 2)),
			fmt.Sprint(roundFloat(initRetractLength-retractLengthDelta*float64(numSegments-1), 2)),
			fmt.Sprint(roundFloat(initRetractSpeed, 0)),
			fmt.Sprint(roundFloat(initRetractSpeed-retractSpeedDelta*float64(numSegments-1), 2)))
		js.Global().Call("saveTextAsFile", fileName, outputGCode)

	}

	return js.ValueOf(nil)
}

func generateLACommand(kFactor float64) string {
	if firmware == 0 {
		return fmt.Sprintf("M900 K%s", fmt.Sprint(roundFloat(kFactor, 3)))
	} else if firmware == 1 {
		return fmt.Sprintf("SET_PRESSURE_ADVANCE ADVANCE=%s", fmt.Sprint(roundFloat(kFactor, 3)))
	} else if firmware == 2 {
		return fmt.Sprintf("M572 D0 S%s", fmt.Sprint(roundFloat(kFactor, 3)))
	}

	return ";no firmware information"
}

func rotateSquareTrajectoryCW(trajectory []Point) []Point {
	rotatedTrajectory := make([]Point, len(trajectory))
	for i := 0; i < len(trajectory); i++ {
		if (i+1)%5 == 0 {
			rotatedTrajectory[i] = trajectory[i-3]
		} else {
			rotatedTrajectory[i] = trajectory[i+1]
		}
	}
	return rotatedTrajectory
}

func generateMove(start, end Point, width float64) []string {
	// create move
	extrude := width > 0
	move := make([]string, 0, 1)
	isMoveOnlyZ := start.X == end.X && start.Y == end.Y

	// if it's travel move, do retraction
	if !extrude && !isMoveOnlyZ {
		move = append(move, generateRetraction())
	}

	// create G1 command
	command := "G1"

	// add X
	if end.X != start.X {
		command = command + fmt.Sprintf(" X%s", fmt.Sprint(roundFloat(end.X, 2)))
	}

	// add Y
	if end.Y != start.Y {
		command = command + fmt.Sprintf(" Y%s", fmt.Sprint(roundFloat(end.Y, 2)))
	}

	// add Z or E. Z move can't be with extrusion
	if end.Z != start.Z {
		command = command + fmt.Sprintf(" Z%s", fmt.Sprint(roundFloat(end.Z, 2)))
	} else if extrude {
		if math.Sqrt(float64(math.Pow((end.X-start.X), 2)+math.Pow((end.Y-start.Y), 2))) > 0.8 {
			newE := currentE + calcExtrusion(start, end, width)
			command = command + fmt.Sprintf(" E%s", fmt.Sprint(roundFloat(newE, 4)))
			currentE = newE
		}
	}

	// add F
	if extrude {
		if currentCoordinates.Z < layerHeight*2 {
			command = command + fmt.Sprintf(" F%s", fmt.Sprint(roundFloat(firstLayerPrintSpeed*60, 0)))
			currentSpeed = firstLayerPrintSpeed
		} else if currentSpeed != printSpeed {
			command = command + fmt.Sprintf(" F%s", fmt.Sprint(roundFloat(printSpeed*60, 0)))
			currentSpeed = printSpeed
		}
	} else {
		if currentSpeed != travelSpeed {
			command = command + fmt.Sprintf(" F%s", fmt.Sprint(roundFloat(travelSpeed*60, 0)))
			currentSpeed = travelSpeed
		}
	}

	// add G1 to move
	move = append(move, command+"\n")
	currentCoordinates = end

	// if there was retraction, than do deretraction
	if !extrude && !isMoveOnlyZ {
		move = append(move, generateDeretraction())
	}

	return move
}

func calcExtrusion(start, end Point, width float64) float64 {
	lineLength := math.Sqrt(float64(math.Pow((end.X-start.X), 2) + math.Pow((end.Y-start.Y), 2)))
	extrusion := width * layerHeight * lineLength * 4 / math.Pi / math.Pow(filamentDiameter, 2)
	return extrusion
}

func generateZigZagTrajectory(towerCenter Point, lineWidth float64) []Point {
	raftWidth := 30.0
	sideLength := raftWidth - lineWidth
	pointsOnOneSide := int(sideLength / (lineWidth * math.Sqrt(2)))
	pointsOnOneSide = pointsOnOneSide - (pointsOnOneSide-1)%2
	pointSpacing := sideLength / float64(pointsOnOneSide-1)
	firstLayerLineWidth = pointSpacing / math.Sqrt(2)

	totalPoints := pointsOnOneSide*4 - 4
	unsortedPoints := make([]Point, totalPoints)

	minX := towerCenter.X - sideLength/2
	minY := towerCenter.Y - sideLength/2
	maxX := towerCenter.X + sideLength/2
	maxY := towerCenter.Y + sideLength/2

	// Generate unsorted slice of points clockwise
	for i := 0; i <= pointsOnOneSide-1; i++ {
		unsortedPoints[i].X = minX + pointSpacing*float64(i)
		unsortedPoints[i].Y = maxY
	}
	for i := 1; i <= pointsOnOneSide-1; i++ {
		unsortedPoints[pointsOnOneSide+i-1].X = maxX
		unsortedPoints[pointsOnOneSide+i-1].Y = maxY - pointSpacing*float64(i)
	}
	for i := 1; i <= pointsOnOneSide-1; i++ {
		unsortedPoints[pointsOnOneSide*2+i-2].X = maxX - pointSpacing*float64(i)
		unsortedPoints[pointsOnOneSide*2+i-2].Y = minY
	}
	for i := 1; i < pointsOnOneSide-1; i++ {
		unsortedPoints[pointsOnOneSide*3+i-3].X = minX
		unsortedPoints[pointsOnOneSide*3+i-3].Y = minY + pointSpacing*float64(i)
	}

	// add Z coordinates

	for i := 1; i < len(unsortedPoints); i++ {
		unsortedPoints[i].Z = currentCoordinates.Z
	}

	// Sort points to make zigzag moves
	trajectory := make([]Point, len(unsortedPoints))

	trajectory[0] = unsortedPoints[0]
	trajectory[1] = unsortedPoints[len(unsortedPoints)-1]
	trajectory[2] = unsortedPoints[1]
	trajectory[3] = unsortedPoints[2]
	for i := 4; i < len(unsortedPoints); i = i + 4 {
		j := int(i / 2)
		trajectory[i] = unsortedPoints[len(unsortedPoints)-j]
		trajectory[i+1] = unsortedPoints[len(unsortedPoints)-j-1]
		trajectory[i+2] = unsortedPoints[j+1]
		trajectory[i+3] = unsortedPoints[j+2]
	}

	for i := 0; i < len(trajectory); i++ {
		trajectory[i].Z = currentCoordinates.Z
	}

	return trajectory
}

func generateSquareTrajectory(squareCenter Point, size float64) []Point {
	// 2----3
	// |    |
	// 1---0,4

	trajectory := make([]Point, 5)
	trajectory[0].X = squareCenter.X + size/2
	trajectory[0].Y = squareCenter.Y - size/2

	trajectory[1].X = squareCenter.X - size/2
	trajectory[1].Y = trajectory[0].Y

	trajectory[2].X = trajectory[1].X
	trajectory[2].Y = squareCenter.Y + size/2

	trajectory[3].X = trajectory[0].X
	trajectory[3].Y = trajectory[2].Y

	trajectory[4] = trajectory[0]

	for i := 0; i < len(trajectory); i++ {
		trajectory[i].Z = currentCoordinates.Z
	}
	return trajectory
}

func generateRetraction() string {
	if retracted {
		fmt.Println("Called retraction, but already retracted")
		return ""
	} else {
		retracted = true
		currentSpeed = retractSpeed
		return fmt.Sprintf("G1 E%s F%s\n", fmt.Sprint(roundFloat(currentE-retractLength, 2)), fmt.Sprint(roundFloat(retractSpeed*60, 0)))
	}
}

func generateDeretraction() string {
	if retracted {
		retracted = false
		currentSpeed = retractSpeed
		return fmt.Sprintf("G1 E%s F%s\n", fmt.Sprint(roundFloat(currentE, 2)), fmt.Sprint(roundFloat(retractSpeed*60, 0)))
	} else {
		fmt.Println("Called deretraction, but not retracted")
		return ""
	}
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func parseInputToFloat(val string) (float64, error) {
	f, err := strconv.ParseFloat(strings.ReplaceAll(val, ",", "."), 64)
	if err != nil {
		println(err.Error())
	}
	return f, err
}

func parseInputToInt(val string) (int, error) {
	f, err := parseInputToFloat(val)
	return int(roundFloat(f, 0)), err
}
