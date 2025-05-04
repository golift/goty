package goty_test

import (
	"time"

	"golift.io/goty"
)

type TestWrapper struct {
	Profile TestLevel1
	Level1  TestLevel1
	TestEndpoint
	EP   *TestEndpoint
	Auth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	Config *goty.Config
}

// Weekdays are cool.
var Weekdays = []goty.Enum{
	{Name: "Sunday", Value: time.Sunday},
	{Name: "Monday", Value: time.Monday},
	{Name: "Tuesday", Value: time.Tuesday},
	{Name: "Wednesday", Value: time.Wednesday},
	{Name: "Thursday", Value: time.Thursday},
	{Name: "Friday", Value: time.Friday},
	{Name: "Saturday", Value: time.Saturday},
}

type TestLevel1 struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type TestEndpoint struct {
	URL    string `json:"url"`
	APIKey string `json:"apiKey"`
}

func ExampleGoty_Print() {
	goty := goty.NewGoty(nil)
	goty.Enums(Weekdays)
	goty.Parse(TestWrapper{})
	goty.Print()
	// Output:
	// /* Auto-generated. DO NOT EDIT. Generator: https://golift.io/goty
	//  * Edit the source code and run goty again to make updates.
	//  */
	//
	// /**
	//  * @see golang: <time.Weekday>
	//  */
	// export enum Weekday {
	//   Sunday    = 0,
	//   Monday    = 1,
	//   Tuesday   = 2,
	//   Wednesday = 3,
	//   Thursday  = 4,
	//   Friday    = 5,
	//   Saturday  = 6,
	// };
	//
	// /**
	//  * @see golang: <golift.io/goty_test.TestWrapper>
	//  */
	// export interface TestWrapper extends TestEndpoint {
	//   Profile: TestLevel1;
	//   Level1: TestLevel1;
	//   EP?: TestEndpoint;
	//   Auth: {
	//     username: string;
	//     password: string;
	//   };
	//   Config?: Config;
	// };
	//
	// /**
	//  * @see golang: <golift.io/goty_test.TestLevel1>
	//  */
	// export interface TestLevel1 {
	//   name: string;
	//   date: Date;
	// };
	//
	// /**
	//  * @see golang: <golift.io/goty_test.TestEndpoint>
	//  */
	// export interface TestEndpoint {
	//   url: string;
	//   apiKey: string;
	// };
	//
	// /**
	//  * @see golang: <golift.io/goty.Config>
	//  */
	// export interface Config {
	//   overrides?: Record<null | any, Override>;
	//   globalOverrides: Override;
	// };
	//
	// /**
	//  * @see golang: <golift.io/goty.Override>
	//  */
	// export interface Override {
	//   type: string;
	//   name: string;
	//   tag: string;
	//   comment: string;
	//   optional: boolean;
	//   keepBadChars: boolean;
	//   keepUnderscores: boolean;
	//   usePkgName: number;
	//   noExport: boolean;
	// };
	//
	// // Packages parsed:
	// //   1. golift.io/goty
	// //   2. golift.io/goty_test
}
