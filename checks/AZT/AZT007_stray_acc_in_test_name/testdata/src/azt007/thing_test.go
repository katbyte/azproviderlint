package azt007_test

import "testing"

// Should be flagged: Acc as an abbreviation inside the step name, in a test or a sequential helper
func TestAccThing_storageAccBehindFireWall(t *testing.T) {} // want `TestAccThing_storageAccBehindFireWall has a second Acc in its name, spell the word out so it is not read as another acceptance marker`
func testAccThing_storageAccBehindFireWall(t *testing.T) {} // want `testAccThing_storageAccBehindFireWall has a second Acc in its name, spell the word out so it is not read as another acceptance marker`
func TestAccThing_withAcc(t *testing.T)                 {} // want `TestAccThing_withAcc has a second Acc in its name, spell the word out so it is not read as another acceptance marker`
func TestAccAcc_basic(t *testing.T)                     {} // want `TestAccAcc_basic has a second Acc in its name, spell the word out so it is not read as another acceptance marker`

// Should NOT be flagged: whole words that start with Acc
func TestAccThing_storageAccountBehindFirewall(t *testing.T) {}
func TestAccThing_publicAccess(t *testing.T)                 {}
func TestAccAccount_basic(t *testing.T)                      {}
func TestAccThing_accelerated(t *testing.T)                  {}

// Should NOT be flagged: not acceptance test names, AZT004 covers those that should be
func TestWebAppAccActiveSlot_basic(t *testing.T) {}
func TestThing_storageAcc(t *testing.T)          {}

type ThingResource struct{}

func (r ThingResource) TestAccThing_storageAccMethod(t *testing.T) {}
