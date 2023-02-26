package main

import (
	"fmt"
	"os"
	"log"
	"bufio"
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

//reads an XYZ file and returns a slice of Point3D
func ReadXYZ(filename string) []Point3D {
	var points []Point3D

	//open the file for reading, checking for errors, and closing after the function
	file, err := os.Open(filename)
	ErrorCheck(err)
	defer file.Close()

	//create a scanner for the file
	scanner := bufio.NewScanner(file)

	//read from the file line by line
	for scanner.Scan() {
		//scan in each point
		point := scanner.Text()
		splitPoint := strings.Split(point, ",")
	
		//store each point into points
		points.append(Point3D{ {splitPoint[0], splitPoint[1], splitPoint[2] })
	}

	//make sure there were no errors while scanning the file
	ErrorCheck(scanner.Err())

	return points
}

//saves a slice of Point3D into an XYZ file
func SaveXYZ(filename string, points []Point3D) {

}

//computes the distance between points p1 and p2
func (p1 *Point3D) GetDistance(p2 *Point3D) float64 {

}

//computes the plane defined by a set of 3 points
func GetPlane(points []Point3D) Plane3D {

}

//computes the number of required RANSAC iterations
func GetNumberOfIterations(confidence float64, percentageOfPointsOnPlane float64) int {

}

//computes the support of a plane in a set of points
func GetSupport(plane Plane3D, points []Point3D, eps float64) Plane3DwSupport {

}

//extracts the points that support the given plane
//and returns them as a slice of points
func GetSupportingPoints(plane Plane3D, points []Point3D, eps float64) []Point3D {

}

//creates a new slice of points in which all points
//belonging to the plane have been removed
func RemovePlane(plane Plane3D, points []Point3D, eps float64) []Point3D {

}

func main() {
	
}