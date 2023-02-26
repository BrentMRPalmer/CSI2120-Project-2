package main

import (
	"fmt"
	"os"
	"log"
	"bufio"
	"strings"
	"strconv"
)

type Point3D struct {
	X float64
	Y float64
	Z float64
}

type Plane3D struct {
	A float64
	B float64
	C float64
	D float64
}

type Plane3DwSupport struct {
	Plane3D
	SupportSize int
}

//used to check for errors
func ErrorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//convert a string to a float
func StringToFloat(s string) float64 {
	x, err := strconv.ParseFloat(s, 64)
	ErrorCheck(err)
	return x
}

//reads an XYZ file and returns a slice of Point3D
func ReadXYZ(filename string) []Point3D {
	var points []Point3D

	//open the file for reading, checking for errors, and closing after the function
	file, err := os.Open(filename)
	ErrorCheck(err)
	defer file.Close()

	//create a scanner for the file and skip the header
	scanner := bufio.NewScanner(file)
	scanner.Scan()


	//read from the file line by line
	for scanner.Scan() {
		//scan in each point
		point := scanner.Text()
		splitPoint := strings.Split(point, "\t")

		//store each point into points
		points = append(points, Point3D{ 
			StringToFloat(splitPoint[0]), 
			StringToFloat(splitPoint[1]),
			StringToFloat(splitPoint[2]),
		})
	}

	//make sure there were no errors while scanning the file
	ErrorCheck(scanner.Err())

	return points
}

//saves a slice of Point3D into an XYZ file
func SaveXYZ(filename string, points []Point3D) {
	//create a file to write to
	file, err := os.Create(filename)
	ErrorCheck(err)
	defer file.Close()
	file.WriteString("x\ty\tz\n")

	//iterate over each point and write each point to the file
	for _, point := range points {
		_, err := file.WriteString(fmt.Sprintf("%f\t%f\t%f\n", point.X, point.Y, point.Z))
		ErrorCheck(err)
	}
	
	//flush buffered data (write data immediately and block function return until written)
	err = file.Sync()
	ErrorCheck(err)
}

// //computes the distance between points p1 and p2
// func (p1 *Point3D) GetDistance(p2 *Point3D) float64 {

// }

// //computes the plane defined by a set of 3 points
// func GetPlane(points []Point3D) Plane3D {

// }

// //computes the number of required RANSAC iterations
// func GetNumberOfIterations(confidence float64, percentageOfPointsOnPlane float64) int {

// }

// //computes the support of a plane in a set of points
// func GetSupport(plane Plane3D, points []Point3D, eps float64) Plane3DwSupport {

// }

// //extracts the points that support the given plane
// //and returns them as a slice of points
// func GetSupportingPoints(plane Plane3D, points []Point3D, eps float64) []Point3D {

// }

// //creates a new slice of points in which all points
// //belonging to the plane have been removed
// func RemovePlane(plane Plane3D, points []Point3D, eps float64) []Point3D {

// }

func PrintPoints(points []Point3D) {
	for i, point := range points {
		fmt.Println(i + 1, point)
	}
}

func main() {
	points := ReadXYZ("PointCloud1.xyz")
	//PrintPoints(points)
	SaveXYZ("newfile.xyz", points)
}