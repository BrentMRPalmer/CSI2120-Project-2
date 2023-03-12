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
	"time"
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
	fmt.Printf("%fx + %fy + %fz = %f\n", plane.A, plane.B, plane.C, plane.D)
}

//Pipeline functions

//STAGE 1
//Randpoint point generator: -> Point3D
//randomly selects a point from the provided slice of Point3D (the input point cloud).
//its output channel transmits instances of Point3D
func RandomPointGenerator(wg *sync.WaitGroup, stop <-chan bool, points []Point3D) <-chan Point3D {
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

//STAGE 2
//Triplet of points generator: Point3D -> [3]Point3D
//it reads Point3D instances from its input channel and accumulates 3 points.
//its output channel transmits arrays of Point3D (composed of 3 points).
func TripletOfPointsGenerator(wg *sync.WaitGroup, stop <-chan bool, randomPoints <-chan Point3D) <-chan [3]Point3D {
	tripletPointStream := make(chan [3]Point3D)
	temp := [3]Point3D{}
	i := 0

	go func() {
		defer func() {wg.Done()} ()
		defer close(tripletPointStream)
		for point := range randomPoints {
			temp[i] = point
			i++
			if (i == 3) {
				select {
					case <- stop:
						return
					case tripletPointStream <- temp:
						temp = [3]Point3D{}
						i = 0
				}
			}
		}
	}()

	return tripletPointStream
}

//STAGE 3
//TakeN: [3]Point3D -> [3]Point3D
//it reads arrays of Point3D and resends them. It automatically stops the pipeline after 
//having received N arrays
func TakeN(wg *sync.WaitGroup, stop chan<- bool, randomTriplet <-chan [3]Point3D, N int) <-chan [3]Point3D {
	tripletPointStream := make(chan [3]Point3D)

	go func() {
		defer func() {
			wg.Done()
			close(stop)
			close(tripletPointStream)
		}()
		for i := 0 ; i < N ; i++ {
			triplet := <- randomTriplet
			tripletPointStream <- triplet
		}
	}()

	return tripletPointStream
}

//STAGE 4 
//Plane estimator: It reads arrays of three Point3D and computes the plane defined by these points
//its ouput channel transmits Plane3D instances describing the computed plane parameters
func PlaneEstimator(wg *sync.WaitGroup, randomTriplet <-chan [3]Point3D) <-chan Plane3D {
	planeStream := make(chan Plane3D)

	go func() {
		defer func() {wg.Done()} ()
		defer close(planeStream)
		for triplet := range randomTriplet {
			plane := GetPlane(triplet[:])
			planeStream <- plane
		}
	}()

	return planeStream
}

//STAGE 5
//Supporting point finder: Plane3D -> Plane3DwSupport
//It counts the number of points in the provided slice of Point3D (the input point cloud)
//that supports the received 3D plane. Its output channel transmits the plane
//parameters and the number of supporting points in a Point3DwSupport instance
func SupportingPointFinder(wg *sync.WaitGroup, randomPlane <-chan Plane3D, points []Point3D, eps float64) <-chan Plane3DwSupport {
	planeSupportStream := make(chan Plane3DwSupport)

	go func() {
		defer func() {wg.Done()} ()
		defer close(planeSupportStream)
		for plane := range randomPlane {
			planeSupport := GetSupport(plane, points[:], eps)
			planeSupportStream <- planeSupport
		}
	}()

	return planeSupportStream
}

// STAGE 6
// Fan In: Plane3DwSupport -> Plane3DwSupport
// It multiplexes the results received from multiple channels into one output channel
func FanIn(wg *sync.WaitGroup, supports[]<-chan Plane3DwSupport) <-chan Plane3DwSupport {
	planeSupportStream := make(chan Plane3DwSupport)
	var fwg sync.WaitGroup
	fwg.Add(len(supports))

	for _, inputPlaneSupportStream := range supports{
		go func(inputPlaneSupportStream <-chan Plane3DwSupport){
			defer wg.Done()
			defer fwg.Done()
			for planeSupport := range inputPlaneSupportStream {
				planeSupportStream <- planeSupport
			}
		}(inputPlaneSupportStream)
	}

	go func(){
		defer wg.Done()
		defer close(planeSupportStream)
		fwg.Wait()
	}()

	return planeSupportStream
}

//STAGE 7
//Dominant Plane Identifer: Plane3DSupport
//It receives Plane3DwSupport instances and keeps in memory the plane with the best support
//received so far. This componenet does not output values, it simply maintains the provided
//*Plane3DwSupport variable
func DominantPointIdentifier(wg *sync.WaitGroup, supportPlanes <-chan Plane3DwSupport, dominantPlane *Plane3DwSupport) {
	for planeToProcess := range supportPlanes{
		if planeToProcess.SupportSize > dominantPlane.SupportSize {
			*dominantPlane = planeToProcess
		}
	}
}

func Pipeline(numThreads int, points []Point3D, bestSupport *Plane3DwSupport, confidence float64, percentageOfPointsOnPlane float64, eps float64, numIterations int) {
	wg := &sync.WaitGroup{}
	stop := make(chan bool)

	wg.Add(1)
	randomPoints := RandomPointGenerator(wg, stop, points)

	wg.Add(1)
	randomTriplets := TripletOfPointsGenerator(wg, stop, randomPoints)

	wg.Add(1)
	nTriplets := TakeN(wg, stop, randomTriplets, numIterations)

	wg.Add(1)
	randomPlane := PlaneEstimator(wg, nTriplets)

	var supportingPointFinders []<-chan Plane3DwSupport
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		supportingPointFinders = append(supportingPointFinders, SupportingPointFinder(wg, randomPlane, points, eps))
	}

	wg.Add(numThreads)
	randomSupportingPointFinder := FanIn(wg, supportingPointFinders)

	wg.Add(1)
	DominantPointIdentifier(wg, randomSupportingPointFinder, bestSupport)

	wg.Wait()
}

func TestTime(points []Point3D, bestSupport *Plane3DwSupport, confidence float64, percentageOfPointsOnPlane float64, eps float64, numIterations int) {
	for i := 1 ; i <= 500; i++ {
		start := time.Now()
		for j := 0 ; j < 200 ; j++ {
			Pipeline(i, points, bestSupport, confidence, percentageOfPointsOnPlane, eps, numIterations)
		}
		end := time.Now()
		elapsed := end.Sub(start)
		fmt.Printf("Average elapsed time with %d threads: %v\n", i, elapsed/200)
	}

}

func main() {
	//read the XYZ file specified as a first argument to your go program and create
	//the corresponding slice of Point3D, composed of the set of points of the XYZ file
	args := os.Args[1:]
	filename := args[0]
	points := ReadXYZ(filename)

	//create a bestSupport variable of type Plane3DwSupport initialized to all 0s
	bestSupport := &Plane3DwSupport{}

	//find the number of iterations required based on the specified confidence and percentage
	//provided as 1st and 2nd arguments arguments for the GetNumberOfIterations function
	confidence := StringToFloat(args[1])
	percentageOfPointsOnPlane := StringToFloat(args[2]) 
	eps := StringToFloat(args[3])
	numIterations := GetNumberOfIterations(confidence, percentageOfPointsOnPlane)

	//create and start the RANSAC find dominant plane pipeline
	//this pipeline automatically stops after the required number of iterations
	//Pipeline(threads, points, bestSupport, confidence, percentageOfPointsOnPlane, eps, numIterations)
	TestTime(points, bestSupport, confidence, percentageOfPointsOnPlane, eps, numIterations)

	SaveXYZ(filename[:len(filename)-4] + "_p.XYZ", GetSupportingPoints(bestSupport.Plane3D, points, eps))
	SaveXYZ(filename[:len(filename)-4] + "_p0.XYZ", RemovePlane(bestSupport.Plane3D, points, eps))

}