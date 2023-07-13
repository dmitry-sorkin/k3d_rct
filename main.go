package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"syscall/js"
)

const filamentDiameter = 1.75
const caliVersion = "v1.5"

var (
	bedX, bedY, lineWidth, firstLayerLineWidth, printSpeed, travelSpeed, layerHeight, initRetractLength, retractLength, retractLengthDelta, currentE, firstLayerPrintSpeed, segmentHeight, towerSpacing, towerWidth, zOffset, initRetractSpeed, retractSpeed, currentSpeed, retractSpeedDelta float64
	hotendTemperature, bedTemperature, numSegments, cooling, flow                                                                                                                                                                                                                             int
	currentCoordinates                                                                                                                                                                                                                                                                        Point
	bedProbe, retracted, delta                                                                                                                                                                                                                                                                bool
	kFactorCommand                                                                                                                                                                                                                                                                            string
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
}

func check() bool {
	errorString := ""
	doc := js.Global().Get("document")
	lang := js.Global().Get("lang")
	doc.Call("getElementById", "resultContainer").Set("innerHTML", "")

	// Fill variables with data from web page
	docBedX, err := parseInputToFloat(doc.Call("getElementById", "bedX").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.bed_size_x.format").String() + "\n"
	} else if docBedX < 100 || docBedX > 1000 {
		errorString = errorString + lang.Call("getString", "error.bed_size_x.small_or_big").String() + "\n"
	} else {
		bedX = docBedX
	}

	docBedY, err := parseInputToFloat(doc.Call("getElementById", "bedY").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.bed_size_y.format").String() + "\n"
	} else if docBedY < 100 || docBedY > 1000 {
		errorString = errorString + lang.Call("getString", "error.bed_size_y.small_or_big").String() + "\n"
	} else {
		bedY = docBedY
	}

	delta = doc.Call("getElementById", "delta").Get("checked").Bool()

	bedProbe = doc.Call("getElementById", "bedProbe").Get("checked").Bool()

	docHotTemp, err := parseInputToInt(doc.Call("getElementById", "hotendTemperature").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.hotend_temp.format").String() + "\n"
	} else if docHotTemp < 150 {
		errorString = errorString + lang.Call("getString", "error.hotend_temp.too_low").String() + "\n"
	} else if docHotTemp > 350 {
		errorString = errorString + lang.Call("getString", "error.hotend_temp.too_high").String() + "\n"
	} else {
		hotendTemperature = docHotTemp
	}

	docBedTemp, err := parseInputToInt(doc.Call("getElementById", "bedTemperature").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.bed_temp.format").String() + err.Error()
	} else if docBedTemp > 150 {
		errorString = errorString + lang.Call("getString", "error.bed_temp.too_high").String() + "\n"
	} else {
		bedTemperature = docBedTemp
	}

	docCooling, err := parseInputToInt(doc.Call("getElementById", "cooling").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.fan_speed.format").String() + "\n"
	} else {
		docCooling = int(float64(docCooling) * 2.55)
		if docCooling < 0 {
			docCooling = 0
		} else if docCooling > 255 {
			docCooling = 255
		}
		cooling = docCooling
	}

	docLineWidth, err := parseInputToFloat(doc.Call("getElementById", "lineWidth").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.line_width.format").String() + "\n"
	} else if docLineWidth < 0.1 || docLineWidth > 2.0 {
		errorString = errorString + lang.Call("getString", "error.line_width.small_or_big").String() + "\n"
	} else {
		lineWidth = docLineWidth
	}

	docFirstLineWidth, err := parseInputToFloat(doc.Call("getElementById", "firstLayerLineWidth").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.first_line_width.format").String() + "\n"
	} else if docFirstLineWidth < 0.1 || docFirstLineWidth > 2.0 {
		errorString = errorString + lang.Call("getString", "error.first_line_width.small_or_big").String() + "\n"
	} else {
		firstLayerLineWidth = docFirstLineWidth
	}

	docLayerHeight, err := parseInputToFloat(doc.Call("getElementById", "layerHeight").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.layer_height.format").String() + "\n"
	} else if docLayerHeight < 0.05 || docLayerHeight > lineWidth*0.75 {
		errorString = errorString + lang.Call("getString", "error.layer_height.small_or_big").String() + "\n"
	} else {
		layerHeight = docLayerHeight
	}

	docPrintSpeed, err := parseInputToFloat(doc.Call("getElementById", "printSpeed").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.print_speed.format").String() + "\n"
	} else if docPrintSpeed < 10 || docPrintSpeed > 1000 {
		errorString = errorString + lang.Call("getString", "error.print_speed.slow_or_fast").String() + "\n"
	} else {
		printSpeed = docPrintSpeed
	}

	docFirstPrintSpeed, err := parseInputToFloat(doc.Call("getElementById", "firstLayerPrintSpeed").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.first_print_speed.format").String() + "\n"
	} else if docFirstPrintSpeed < 10 || docFirstPrintSpeed > 1000 {
		errorString = errorString + lang.Call("getString", "error.first_print_speed.slow_or_fast").String() + "\n"
	} else {
		firstLayerPrintSpeed = docFirstPrintSpeed
	}

	docTravelSpeed, err := parseInputToFloat(doc.Call("getElementById", "travelSpeed").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.travel_speed.format").String() + "\n"
	} else if docTravelSpeed < 10 || docTravelSpeed > 1000 {
		errorString = errorString + lang.Call("getString", "error.travel_speed.slow_or_fast").String() + "\n"
	} else {
		travelSpeed = docTravelSpeed
	}

	docNumSegments, err := parseInputToInt(doc.Call("getElementById", "numSegments").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.num_segments.format").String() + "\n"
	} else if docNumSegments < 2 || docNumSegments > 100 {
		errorString = errorString + lang.Call("getString", "error.num_segments.slow_or_fast").String() + "\n"
	} else {
		numSegments = docNumSegments
	}

	docInitRetractLength, err := parseInputToFloat(doc.Call("getElementById", "initRetractLength").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.init_retract_length.format").String() + "\n"
	} else if docInitRetractLength < 0 || docInitRetractLength > 20 {
		errorString = errorString + lang.Call("getString", "error.init_retract_length.small_or_big").String() + "\n"
	} else {
		retractLength = docInitRetractLength
		initRetractLength = docInitRetractLength
	}

	docEndRetractLength, err := parseInputToFloat(doc.Call("getElementById", "endRetractLength").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.end_retract_length.format").String() + "\n"
	} else if docEndRetractLength < 0 || docEndRetractLength > 20 {
		errorString = errorString + lang.Call("getString", "error.end_retract_length.small_or_big").String() + "\n"
	} else {
		retractLengthDelta = (docInitRetractLength - docEndRetractLength) / float64(numSegments-1)
	}

	docRetractSpeed, err := parseInputToFloat(doc.Call("getElementById", "initRetractSpeed").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.init_retract_speed.format").String() + "\n"
	} else if docRetractSpeed < 5 || docRetractSpeed > 150 {
		errorString = errorString + lang.Call("getString", "error.init_retract_speed.slow_or_fast").String() + "\n"
	} else {
		retractSpeed = docRetractSpeed
		initRetractSpeed = docRetractSpeed
	}

	docEndRetractSpeed, err := parseInputToFloat(doc.Call("getElementById", "endRetractSpeed").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.end_retract_speed.format").String() + "\n"
	} else if docEndRetractSpeed < 5 || docEndRetractSpeed > 150 {
		errorString = errorString + lang.Call("getString", "error.end_retract_speed.slow_or_fast").String() + "\n"
	} else {
		retractSpeedDelta = (docRetractSpeed - docEndRetractSpeed) / float64(numSegments-1)
	}

	docSegmentHeight, err := parseInputToFloat(doc.Call("getElementById", "segmentHeight").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.segment_height.format").String() + "\n"
	} else if docSegmentHeight < 0.5 || docSegmentHeight > 20 {
		errorString = errorString + lang.Call("getString", "error.segment_height.small_or_big").String() + "\n"
	} else {
		segmentHeight = docSegmentHeight
	}

	kFactorCommand = doc.Call("getElementById", "kFactor").Get("value").String() + "\n"

	docTowerSpacing, err := parseInputToFloat(doc.Call("getElementById", "towerSpacing").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.tower_spacing.format").String() + "\n"
	} else if docTowerSpacing < 40 {
		errorString = errorString + lang.Call("getString", "error.tower_spacing.too_small").String() + "\n"
	} else if docTowerSpacing > bedX-40.0 {
		errorString = errorString + lang.Call("getString", "error.tower_spacing.too_big").String() + "\n"
	} else {
		towerSpacing = docTowerSpacing
	}

	docZOffset, err := parseInputToFloat(doc.Call("getElementById", "zOffset").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.z_offset.format").String() + "\n"
	} else if docZOffset < -layerHeight || docZOffset > layerHeight {
		errorString = errorString + lang.Call("getString", "error.z_offset.too_big").String() + "\n"
	} else {
		zOffset = docZOffset
	}
	
	docFlow, err := parseInputToInt(doc.Call("getElementById", "flow").Get("value").String())
	if err != nil {
		errorString = errorString + lang.Call("getString", "error.flow.format").String() + "\n"
	} else if docFlow < 50 || docFlow > 150 {
		errorString = errorString + lang.Call("getString", "error.flow.low_or_high").String() + "\n"
	} else {
		flow = docFlow
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

func generate(this js.Value, i []js.Value) interface{} {
	// check and initialize variables
	if check() {
		lang := js.Global().Get("lang")
		var segmentStr = lang.Call("getString", "generator.segment").String()
		
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
			caliVersion,
			"\n",
			"; Written by Dmitry Sorkin @ http://k3d.tech/\n",
			"; and Kekht\n",
			fmt.Sprintf(";Bedsize: %f:%f\n", bedX, bedY),
			fmt.Sprintf(";Temp: %d/%d\n", hotendTemperature, bedTemperature),
			fmt.Sprintf(";Width: %f-%f\n", lineWidth, firstLayerLineWidth),
			fmt.Sprintf(";Layer height: %f\n", layerHeight),
			fmt.Sprintf(";Retract length: %f, -%f/segment\n", retractLength, retractLengthDelta),
			fmt.Sprintf(";Segments: %dx%f mm\n", numSegments, segmentHeight),
			caliParams,
			kFactorCommand,
			fmt.Sprintf("M190 S%d\n", bedTemperature),
			fmt.Sprintf("M109 S%d\n", hotendTemperature),
			"G28\n")
		if bedProbe {
			gcode = append(gcode, "G29\n")
		}
		gcode = append(gcode, "G92 E0\n",
			"G90\n",
			"M82\n",
			fmt.Sprintf("M106 S%d\n", int(cooling/3)),
			fmt.Sprintf("M221 S%d\n", flow))

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
		gcode = append(gcode, ";end gcode\n",
			"M104 S0\n",
			"M140 S0\n",
			"M106 S0\n",
			fmt.Sprintf("G1 Z%f F600\n", currentCoordinates.Z+5),
			fmt.Sprintf("G1 X%f Y%f F3000\n", bedCenter.X, bedCenter.Y),
			"M84")

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
