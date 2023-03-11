package main

import (
	"fmt"
	"os"
	"log"
	"bufio"
	"strings"
	"strconv"
	"math"
	"sync"
	"math/rand"
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

//computes the distance between points p1 and p2
func (p1 *Point3D) GetDistance(p2 *Point3D) float64 {
	return math.Sqrt( (p2.X - p1.X)*(p2.X - p1.X) + (p2.Y - p1.Y)*(p2.Y - p1.Y) + (p2.Z-p1.Z)*(p2.Z-p1.Z) )
}

//computes the distance between a point and a plane
func (plane *Plane3D) GetDistance(point *Point3D) float64{
	return math.Abs( plane.A*point.X + plane.B*point.Y + plane.C*point.Z - plane.D) / math.Sqrt(plane.A*plane.A + plane.B*plane.B + plane.C*plane.C)
}

//computes the plane defined by a set of 3 points
func GetPlane(points []Point3D) Plane3D {
	//Assuming plane in Ax + Bx + Cz = D form

	//calculate the first vector (p2p1)
	x1 := points[1].X - points[0].X
	y1 := points[1].Y - points[0].Y
	z1 := points[1].Z - points[0].Z

	//calculate the second vector (p3p1)
	x2 := points[2].X - points[0].X
	y2 := points[2].Y - points[0].Y
	z2 := points[2].Z - points[0].Z

	//calculate the normal of the two vectors
	a := y1 * z2 - z1 * y2
	b := z1 * x2 - x1 * z2
	c := x1 * y2 - y1 * x2

	//use a point to calculate the value of d
	d := a * points[0].X + b * points[0].Y + c * points[0].Z

	return Plane3D{A: a, B: b, C: c, D: d}
}

//computes the number of required RANSAC iterations
func GetNumberOfIterations(confidence float64, percentageOfPointsOnPlane float64) int {
	// l - |a*x + b*y + c*z - d| / sqrt(a^2, b^2, c^2)
	return int( math.Ceil( math.Log10(1 - confidence) / math.Log10(1 - math.Pow(percentageOfPointsOnPlane, 3) ) ) )
}

// //computes the support of a plane in a set of points
func GetSupport(plane Plane3D, points []Point3D, eps float64) Plane3DwSupport {
	support := Plane3DwSupport { plane, 0 }

	for _, point := range points{
		if(plane.GetDistance(&point) < eps){
			support.SupportSize++;
		}
	}
	return support;
}

//extracts the points that support the given plane
//and returns them as a slice of points
func GetSupportingPoints(plane Plane3D, points []Point3D, eps float64) []Point3D {
	var supportingPoints []Point3D
	for _, point := range points{
		if(plane.GetDistance(&point) < eps){
			supportingPoints = append(supportingPoints, point)
		}
	}
	return supportingPoints
}

//creates a new slice of points in which all points
//belonging to the plane have been removed
func RemovePlane(plane Plane3D, points []Point3D, eps float64) []Point3D {
	var leftoverPoints []Point3D
	for _, point := range points{
		if(plane.GetDistance(&point) >= eps){
			leftoverPoints = append(leftoverPoints, point)
		}
	}
	return leftoverPoints
}

func PrintPoints(points []Point3D) {
	for i, point := range points {
		fmt.Println(i + 1, point)
	}
}

//prints out a plane equation in Ax + By + Cz = D format
func (plane *Plane3D) Print() {
	fmt.Printf("%fx + %fy + %fz = %f", plane.A, plane.B, plane.C, plane.D)
}

//Pipeline functions

//Randpoint point generator: -> Point3D
//randomly selects a point from the provided slice of Point3D (the input point cloud).
//its output channel transmits instances of Point3D
func randomPointGenerator(wg *sync.WaitGroup, stop <-chan bool, points []Point3D) <-chan Point3D {
	pointStream := make(chan Point3D)

	go func() {
		defer func() {wg.Done()} ()
		defer close(pointStream)
		for {
			select {
				case <- stop:
					return
				case pointStream <- points[rand.Intn(len(points))]:
			}
		}
	}()

	return pointStream
}


func main() {
	//read the XYZ file specified as a first argument to your go program and create
	//the corresponding slice of Point3D, composed of the set of points of the XYZ file
	args := os.Args[1:]
	points := ReadXYZ(args[0])

	//create a bestSupport variable of type Plane3DwSupport initialized to all 0s
	//var bestSupport Plane3DwSupport

	//find the number of iterations required based on the specified confidence and percentage
	//provided as 1st and 2nd arguments arguments for the GetNumberOfIterations function
	confidence := StringToFloat(args[1])
	percentageOfPointsOnPlane := StringToFloat(args[2]) 
	numIterations := GetNumberOfIterations(confidence, percentageOfPointsOnPlane)
	fmt.Printf("%d\n", numIterations)

	//create and start the RANSAC find dominant plane pipeline
	//this pipeline automatically stops after the required number of iterations
	wg := &sync.WaitGroup{}
	stop := make(chan bool)

	wg.Add(1)
	randomPoints := randomPointGenerator(wg, stop, points)

	for i := 0 ; i < 10 ; i++ {
		point := <- randomPoints
		fmt.Printf("%v\n", point)
	}
	stop <- true
	close(stop)
	wg.Wait()
}