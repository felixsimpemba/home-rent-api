package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/felixsimpemba/home-rent-api/internal/config"
	"github.com/felixsimpemba/home-rent-api/internal/models"
)

func main() {
	// 1. Load config
	cfg := config.LoadConfig()

	// 2. Connect to DB
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBDatabase,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	log.Println("Connected to database for seeding...")

	// 3. Clear existing properties and seed users to clean up sandbox
	db.Exec("DELETE FROM amenities")
	db.Exec("DELETE FROM property_images")
	db.Exec("DELETE FROM properties")
	db.Exec("DELETE FROM users WHERE email IN ('landlord@homerent.zm', 'agent@homerent.zm')")

	// 4. Create mock host users
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Password@123"), bcrypt.DefaultCost)
	
	landlord := models.User{
		ID:            "usr_landlord_seed",
		FirstName:     "Mwansa",
		LastName:      "Chilufya",
		Email:         "landlord@homerent.zm",
		Phone:         "+260977112233",
		Password:      string(hashedPassword),
		Role:          "landlord",
		EmailVerified: true,
	}
	db.Create(&landlord)

	agent := models.User{
		ID:            "usr_agent_seed",
		FirstName:     "Chipo",
		LastName:      "Phiri",
		Email:         "agent@homerent.zm",
		Phone:         "+260966445566",
		Password:      string(hashedPassword),
		Role:          "agent",
		EmailVerified: true,
	}
	db.Create(&agent)

	// 5. Seed properties (Status: "active" so they display in search)
	properties := []models.Property{
		{
			ID:            "prop_seed_1",
			Title:         "Modern 3 Bedroom House",
			Description:   "A beautiful and secure 3-bedroom house in the heart of Ibex Hill, Lusaka. Features a spacious private yard, modern kitchen fittings, and reliable water supply with a borehole.",
			PropertyType:  "house",
			ListingType:   "rent",
			Price:         8500,
			Deposit:       8500,
			Bedrooms:      3,
			Bathrooms:     2,
			ParkingSpaces: 2,
			Province:      "Lusaka",
			District:      "Lusaka",
			Area:          "Ibex Hill",
			Address:       "Plot 45, Twin Palm Road, Ibex Hill",
			Latitude:      -15.4180,
			Longitude:     28.3560,
			Status:        "active",
			Featured:      true,
			OwnerID:       landlord.ID,
		},
		{
			ID:            "prop_seed_2",
			Title:         "Premium 2 Bedroom Apartment",
			Description:   "Executive 2-bedroom fully furnished apartment located in Woodlands. Comes with 24/7 security guard services, backup generator power, paved yard, and proximity to shopping malls.",
			PropertyType:  "apartment",
			ListingType:   "rent",
			Price:         12000,
			Deposit:       12000,
			Bedrooms:      2,
			Bathrooms:     2,
			ParkingSpaces: 1,
			Province:      "Lusaka",
			District:      "Lusaka",
			Area:          "Woodlands",
			Address:       "Woodlands Extension, off Leopards Hill Road",
			Latitude:      -15.4430,
			Longitude:     28.3310,
			Status:        "active",
			Featured:      true,
			OwnerID:       agent.ID,
		},
		{
			ID:            "prop_seed_3",
			Title:         "Cozy Studio Apartment near UNZA",
			Description:   "Clean and quiet bachelor/studio apartment ideal for students or working professionals. Water and trash collection included in rent. Fully fenced and secure.",
			PropertyType:  "studio",
			ListingType:   "rent",
			Price:         3500,
			Deposit:       3500,
			Bedrooms:      1,
			Bathrooms:     1,
			ParkingSpaces: 1,
			Province:      "Lusaka",
			District:      "Lusaka",
			Area:          "Handsworth",
			Address:       "Handsworth Court, near Great East Road",
			Latitude:      -15.3930,
			Longitude:     28.3280,
			Status:        "active",
			Featured:      false,
			OwnerID:       landlord.ID,
		},
		{
			ID:            "prop_seed_4",
			Title:         "Spacious 4 Bedroom Family Villa",
			Description:   "Elegant 4-bedroom villa with master self-contained, swimming pool, manicured gardens, electric fence, and double automated garage. Located in a quiet, upmarket neighborhood of Riverside, Kitwe.",
			PropertyType:  "house",
			ListingType:   "rent",
			Price:         15000,
			Deposit:       15000,
			Bedrooms:      4,
			Bathrooms:     3,
			ParkingSpaces: 3,
			Province:      "Copperbelt",
			District:      "Kitwe",
			Area:          "Riverside",
			Address:       "Jambo Drive, Riverside",
			Latitude:      -12.7910,
			Longitude:     28.2380,
			Status:        "active",
			Featured:      true,
			OwnerID:       agent.ID,
		},
	}

	for _, p := range properties {
		if err := db.Create(&p).Error; err != nil {
			log.Printf("Failed to create property %s: %v", p.Title, err)
			continue
		}

		// Seed images
		images := []models.PropertyImage{
			{
				ID:         "img_" + uuid.NewString(),
				PropertyID: p.ID,
				URL:        "https://images.unsplash.com/photo-1564013799919-ab600027ffc6?auto=format&fit=crop&w=800&q=80",
				Caption:    "Exterior View",
				SortOrder:  0,
			},
			{
				ID:         "img_" + uuid.NewString(),
				PropertyID: p.ID,
				URL:        "https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=800&q=80",
				Caption:    "Living Room",
				SortOrder:  1,
			},
		}
		for _, img := range images {
			db.Create(&img)
		}

		// Seed amenities
		amenitiesList := []string{"borehole", "electricity", "paved yard", "security guard", "generator", "pool"}
		for _, name := range amenitiesList {
			db.Create(&models.Amenity{
				ID:         "amen_" + uuid.NewString(),
				PropertyID: p.ID,
				Name:       name,
			})
		}
	}

	log.Println("✓ Seed completed successfully!")
}
