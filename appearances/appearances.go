package appearances

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/opentibiabr/client-editor/appearances/gen"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func checkUnknown(label string, msg protoreflect.ProtoMessage) {
	unknown := msg.ProtoReflect().GetUnknown()
	for len(unknown) > 0 {
		num, typ, n := protowire.ConsumeField(unknown)
		fmt.Printf("[%s] num: %d, typ: %d, n: %d\n", label, num, typ, n)
		unknown = unknown[n:]
	}
}

func Appearances(appearancesPath, outputAppearancesPath string) {
	// Read the binary data from the appearances.dat file
	data, err := ioutil.ReadFile(appearancesPath)
	if err != nil {
		log.Fatalf("Failed to read the file: %v", err)
	}

	// Initialize the Appearances message
	appearancesData := &gen.Appearances{}

	// Unmarshal the binary data into the Appearances message
	if err := proto.Unmarshal(data, appearancesData); err != nil {
		log.Fatalf("Failed to unmarshal the data: %v", err)
	}
	checkUnknown("Appearances", appearancesData)
	allAppearances := append(appearancesData.Object, appearancesData.Effect...)
	allAppearances = append(allAppearances, appearancesData.Missile...)
	allAppearances = append(allAppearances, appearancesData.Outfit...)
	for _, appearance := range allAppearances {
		checkUnknown("Appearance", appearance)
		checkUnknown("Flags", appearance.Flags)
		checkUnknown("Flags.Automap", appearance.Flags.Automap)
		checkUnknown("Flags.Bank", appearance.Flags.Bank)
		checkUnknown("Flags.Hook", appearance.Flags.Hook)
		checkUnknown("Flags.Market", appearance.Flags.Market)
		checkUnknown("Flags.Shift", appearance.Flags.Shift)
		checkUnknown("Flags.Lenshelp", appearance.Flags.Lenshelp)
		checkUnknown("Flags.Light", appearance.Flags.Light)
		checkUnknown("Flags.Write", appearance.Flags.Write)
		checkUnknown("Flags.WriteOnce", appearance.Flags.WriteOnce)
		checkUnknown("Flags.Height", appearance.Flags.Height)
		checkUnknown("Flags.Clothes", appearance.Flags.Clothes)
		checkUnknown("Flags.DefaultAction", appearance.Flags.DefaultAction)
		checkUnknown("Flags.Changedtoexpire", appearance.Flags.Changedtoexpire)
		checkUnknown("Flags.Cyclopediaitem", appearance.Flags.Cyclopediaitem)
		checkUnknown("Flags.Upgradeclassification", appearance.Flags.Upgradeclassification)

		for _, npcSaleData := range appearance.Flags.Npcsaledata {
			checkUnknown("Flags.NpcSaleData", npcSaleData)
		}
		for _, frameGroup := range appearance.FrameGroup {
			checkUnknown("FrameGroup", frameGroup)
			checkUnknown("SpriteInfo", frameGroup.SpriteInfo)
			checkUnknown("SpriteInfo.Animation", frameGroup.SpriteInfo.Animation)
			for _, box := range frameGroup.SpriteInfo.BoundingBoxPerDirection {
				checkUnknown("BoundingBoxPerDirection", box)
			}
			if frameGroup.SpriteInfo.Animation != nil {
				for _, sprite := range frameGroup.SpriteInfo.Animation.SpritePhase {
					checkUnknown("SpritePhase", sprite)
				}
			}
		}
	}

	checkUnknown("SpecialMeaningAppearanceIds", appearancesData.SpecialMeaningAppearanceIds)

	edits := map[uint32]*gen.AppearanceFlags{}
	for _, rawEdit := range viper.Get("edit").([]interface{}) {
		edit := rawEdit.(map[string]interface{})
		id, err := strconv.Atoi(edit["id"].(string))
		if err != nil {
			log.Fatalf("Failed to parse id: %v", err)
		}
		delete(edit, "id")
		jsonEdit, err := json.Marshal(edit)
		if err != nil {
			log.Fatalf("Failed to marshal edit: %v", err)
		}
		newEdit := &gen.AppearanceFlags{}
		err = json.Unmarshal(jsonEdit, newEdit)
		if err != nil {
			log.Fatalf("Failed to unmarshal edit: %v", err)
		}
		edits[uint32(id)] = newEdit
	}

	for _, appearance := range appearancesData.Object {
		if appearance.Id == nil {
			continue
		}
		id := appearance.GetId()
		if edit, ok := edits[id]; ok {
			edit.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
				if v.IsValid() {
					appearance.Flags.ProtoReflect().Set(fd, v)
				}
				return true
			})
		}
	}

	out, err := proto.Marshal(appearancesData)
	if err != nil {
		log.Fatalf("Failed to marshal the data: %v", err)
	}
	ioutil.WriteFile(outputAppearancesPath, out, os.ModePerm)
}
