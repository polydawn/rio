package rio

import (
	"github.com/polydawn/refmt/obj/atlas"

	"go.polydawn.net/timeless-api"
)

var Atlas = atlas.MustBuild(
	atlas.BuildEntry(Event{}).StructMap().Autogenerate().Complete(),
	atlas.BuildEntry(Event_Progress{}).StructMap().Autogenerate().Complete(),
	atlas.BuildEntry(Event_Result{}).StructMap().Autogenerate().Complete(),
	atlas.BuildEntry(Error{}).StructMap().Autogenerate().Complete(),
	api.WareID_AtlasEntry,
)
