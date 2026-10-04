package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	BaseDeliveryRate           = 0.05
	MaxBatteryLevel    float64 = 100.0
	EmergencyThreshold float64 = 20.0
	BufferCapacity             = 5
)

type RobotType string

const (
	Drone       RobotType = "Drone"
	Wheeled     RobotType = "Wheeled"
	HeavyLifter RobotType = "HeavyLifter"
)

var robotTypes = []RobotType{
	Drone,
	Wheeled,
	HeavyLifter,
}

type Product struct {
	ID     int
	Name   string
	Weight float64
}

type Robot struct {
	ID          int
	Model       string
	Type        RobotType
	Battery     float64
	IsAvailable bool
}

type Mover interface {
	Move(distance float64) error
}

func (r *Robot) Move(distance float64) error {
	var consumptionRate float64

	switch r.Type {
	case Drone:
		consumptionRate = 1.5

	case Wheeled:
		consumptionRate = 0.8

	case HeavyLifter:
		consumptionRate = 2.5

	default:
		return fmt.Errorf("unknown robot type")
	}

	requiredEnergy := distance * (consumptionRate + BaseDeliveryRate)

	if r.Battery < requiredEnergy {
		return fmt.Errorf("not enough battery")
	}

	r.Battery = r.Battery - requiredEnergy

	return nil
}

func Recharge(robot *Robot) {
	robot.Battery = MaxBatteryLevel
}

func canCarry(robot *Robot, product Product) bool {
	var maxWeight float64

	switch robot.Type {
	case Drone:
		maxWeight = 2.0

	case Wheeled:
		maxWeight = 5.0

	case HeavyLifter:
		maxWeight = 20.0

	default:
		return false
	}

	return product.Weight <= maxWeight
}

func addProductToOrder(
	orderBuffer *[BufferCapacity]Product,
	orderCount *int,
	product Product,
) error {

	if *orderCount >= BufferCapacity {
		return fmt.Errorf("order buffer is full")
	}

	orderBuffer[*orderCount] = product
	(*orderCount)++

	return nil
}

func DeliveryTask(
	robot *Robot,
	product Product,
	distance float64,
	reportChan chan<- string,
) {
	time.Sleep(time.Duration(distance*100) * time.Millisecond)

	err := robot.Move(distance)

	if err != nil {
		status := strings.ToUpper("ошибка")

		message := fmt.Sprintf(
			"Robot %s failed to deliver %s. Status: %s",
			robot.Model,
			product.Name,
			status,
		)

		message = strings.TrimSpace(message)

		reportChan <- message
		return
	}

	status := strings.ToUpper("успешно")

	message := fmt.Sprintf(
		"Robot %s delivered %s. Status: %s. Battery: %.2f",
		robot.Model,
		product.Name,
		status,
		robot.Battery,
	)

	message = strings.TrimSpace(message)

	reportChan <- message
}

func main() {
	inventory := map[string]int{
		"Product1": 10,
		"Product2": 5,
		"Product3": 20,
		"Product4": 15,
		"Product5": 8,
	}

	productWeights := map[string]float64{
		"Product1": 1.0,
		"Product2": 2.5,
		"Product3": 3.0,
		"Product4": 2.0,
		"Product5": 1.5,
	}

	robots := []Robot{
		{1, "DroneX", Drone, 100.0, true},
		{2, "WheeledY", Wheeled, 100.0, true},
		{3, "HeavyZ", HeavyLifter, 100.0, true},
		{4, "DroneA", Drone, 100.0, true},
	}

	orderBuffer := [BufferCapacity]Product{}

	orderCount := 0
	nextProductID := 1

	for {
		fmt.Println()
		fmt.Println("===== Nexus-Logistics =====")
		fmt.Println("1. Add Product to Order")
		fmt.Println("2. Show Warehouse")
		fmt.Println("3. Start Delivery")
		fmt.Println("4. Exit")

		var choice int

		fmt.Print("Enter your choice: ")
		fmt.Scan(&choice)

		switch choice {

		case 1:
			if orderCount >= BufferCapacity {
				fmt.Println("Order buffer is full.")
				continue
			}

			var productName string

			fmt.Print("Enter product name: ")
			fmt.Scan(&productName)

			quantity, exists := inventory[productName]

			if !exists || quantity <= 0 {
				fmt.Println("Product not available in inventory.")
				continue
			}

			weight, exists := productWeights[productName]

			if !exists {
				fmt.Println("Product weight information not available.")
				continue
			}

			product := Product{
				ID:     nextProductID,
				Name:   productName,
				Weight: weight,
			}

			err := addProductToOrder(
				&orderBuffer,
				&orderCount,
				product,
			)

			if err != nil {
				fmt.Println(err)
				continue
			}

			inventory[productName]--

			nextProductID++

			fmt.Println("Product added to order.")
			fmt.Println("Order count:", orderCount)

		case 2:
			fmt.Println()
			fmt.Println("Inventory:")

			for name, quantity := range inventory {
				fmt.Println(name, "-", quantity)
			}

			fmt.Println()
			fmt.Println("Robots:")

			for _, robot := range robots {
				var status string

				if robot.IsAvailable {
					status = "available"
				} else {
					status = "busy"
				}

				fmt.Println(
					"ID:", robot.ID,
					"Model:", robot.Model,
					"Type:", robot.Type,
					"Battery:", robot.Battery,
					"Status:", status,
				)
			}

			fmt.Println()
			fmt.Println("Order buffer:")

			if orderCount == 0 {
				fmt.Println("Order buffer is empty.")
			}

			for i := 0; i < orderCount; i++ {
				fmt.Println(
					"ID:", orderBuffer[i].ID,
					"Name:", orderBuffer[i].Name,
					"Weight:", orderBuffer[i].Weight,
				)
			}

		case 3:
			if orderCount == 0 {
				fmt.Println("No orders to deliver.")
				continue
			}

			fmt.Println("Starting delivery...")

			reportChan := make(chan string)

			activeJobs := 0

			var usedRobots []*Robot

			for i := 0; i < orderCount; i++ {
				product := orderBuffer[i]

				var selectedRobot *Robot

				for j := 0; j < len(robots); j++ {

					if robots[j].IsAvailable &&
						robots[j].Battery > EmergencyThreshold &&
						canCarry(&robots[j], product) {

						selectedRobot = &robots[j]
						break
					}
				}

				if selectedRobot == nil {
					fmt.Println(
						"No available robot for",
						product.Name,
					)

					continue
				}

				var distance float64

				fmt.Print(
					"Enter distance for ",
					product.Name,
					": ",
				)

				fmt.Scan(&distance)

				if distance <= 0 {
					fmt.Println("Distance must be greater than 0.")
					continue
				}

				selectedRobot.IsAvailable = false

				usedRobots = append(
					usedRobots,
					selectedRobot,
				)

				go DeliveryTask(
					selectedRobot,
					product,
					distance,
					reportChan,
				)

				activeJobs++
			}

			for i := 0; i < activeJobs; i++ {
				message := <-reportChan

				fmt.Println(message)
			}

			for _, robot := range usedRobots {
				robot.IsAvailable = true

				if robot.Battery < EmergencyThreshold {
					fmt.Println(
						"Robot",
						robot.Model,
						"has low battery. Recharging...",
					)

					Recharge(robot)
				}
			}

			close(reportChan)

			orderCount = 0

			fmt.Println("Delivery process finished.")

		case 4:
			fmt.Println("Program finished.")
			return

		default:
			fmt.Println("Wrong option.")
		}
	}
}
