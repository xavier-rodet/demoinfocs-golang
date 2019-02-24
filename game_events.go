package demoinfocs

import (
	"fmt"
	"strconv"

	r3 "github.com/golang/geo/r3"

	common "github.com/markus-wa/demoinfocs-golang/common"
	events "github.com/markus-wa/demoinfocs-golang/events"
	msg "github.com/markus-wa/demoinfocs-golang/msg"
)

func (p *Parser) handleGameEventList(gel *msg.CSVCMsg_GameEventList) {
	defer func() {
		p.setError(recoverFromUnexpectedEOF(recover()))
	}()

	p.gameEventDescs = make(map[int32]*msg.CSVCMsg_GameEventListDescriptorT)
	for _, d := range gel.GetDescriptors() {
		p.gameEventDescs[d.GetEventid()] = d
	}
}

func (p *Parser) storeGameEvent(ge *msg.CSVCMsg_GameEvent) {
	p.currentGameEvents = append(p.currentGameEvents, ge)
}

func (p *Parser) handleGameEvent(ge *msg.CSVCMsg_GameEvent) {
	defer func() {
		p.setError(recoverFromUnexpectedEOF(recover()))
	}()

	if p.gameEventDescs == nil {
		p.eventDispatcher.Dispatch(events.ParserWarn{Message: "Received GameEvent but event descriptors are missing"})
		return
	}

	desc := p.gameEventDescs[ge.Eventid]

	debugGameEvent(desc, ge)

	if handler, eventKnown := gameEventNameToHandler[desc.Name]; eventKnown {
		handler(p, desc, ge)
	} else {
		handleUnknownEvent(p, desc, ge)
	}
}

var gameEventNameToHandler = map[string]gameEventHandler{
	"round_start":                     handleRoundStart,                 // Round started
	"cs_win_panel_match":              handleCsWinPanelMatch,            // Not sure, maybe match end event???
	"round_announce_final":            handleRoundAnnounceFinal,         // 30th round for normal de_, not necessarily matchpoint
	"round_announce_last_round_half":  handleRoundAnnounceLastRoundHalf, // Last round of the half
	"round_end":                       handleRoundEnd,                   // Round ended and the winner was announced
	"round_officially_ended":          handleRoundOfficiallyEnded,       // The event after which you get teleported to the spawn (=> You can still walk around between round_end and this event)
	"round_mvp":                       handleRoundMVP,                   // Round MVP was announced
	"bot_takeover":                    handleBotTakeover,                // Bot got taken over
	"begin_new_match":                 handleBeginNewMatch,              // Match started
	"round_freeze_end":                handleRoundFreezeEnd,             // Round start freeze ended
	"player_footstep":                 handlePlayerFootstep,             // Footstep sound
	"player_jump":                     handlePlayerJump,                 // Player jumped
	"weapon_fire":                     handleWeaponFire,                 // Weapon was fired
	"player_death":                    handlePlayerDeath,                // Player died
	"player_hurt":                     handlePlayerHurt,                 // Player got hurt
	"player_blind":                    handlePlayerBlind,                // Player got blinded by a flash
	"flashbang_detonate":              handleFlashbangDetonate,          // Flash exploded
	"hegrenade_detonate":              handleHegranadeDetonate,          // HE exploded
	"decoy_started":                   handleDecoyStarted,               // Decoy started
	"decoy_detonate":                  handleDecoyDetonate,              // Decoy exploded/expired
	"smokegrenade_detonate":           handleSmokegrenadeDetonate,       // Smoke popped
	"smokegrenade_expired":            handleSmokegrenadeExpired,        // Smoke expired
	"inferno_startburn":               handleInfernoStartburn,           // Incendiary exploded/started
	"inferno_expire":                  handleInfernoExpire,              // Incendiary expired
	"player_connect":                  handlePlayerConnect,              // Bot connected or player reconnected, players normally come in via string tables & data tables
	"player_disconnect":               handlePlayerDisconnect,           // Player disconnected (kicked, quit, timed out etc.)
	"player_team":                     handlePlayerTeam,                 // Player changed team
	"bomb_beginplant":                 handleBombBeginplant,             // Plant started
	"bomb_planted":                    handleBombPlanted,                // Plant finished
	"bomb_defused":                    handleBombDefused,                // Defuse finished
	"bomb_exploded":                   handleBombExploded,               // Bomb exploded
	"bomb_begindefuse":                handleBombBegindefuse,            // Defuse started
	"item_equip":                      handleItemEquip,                  // Equipped, I think
	"item_pickup":                     handleItemPickup,                 // Picked up or bought?
	"item_remove":                     handleItemRemove,                 // Dropped?
	"bomb_dropped":                    handleBombDropped,                // Bomb dropped
	"bomb_pickup":                     handleBombPickup,                 // Bomb picked up
	"player_connect_full":             handleGenericEvent,               // Connecting finished
	"player_falldamage":               handleGenericEvent,               // Falldamage
	"weapon_zoom":                     handleGenericEvent,               // Zooming in
	"weapon_reload":                   handleGenericEvent,               // Weapon reloaded
	"round_time_warning":              handleGenericEvent,               // Round time warning
	"round_announce_match_point":      handleGenericEvent,               // Match point announcement
	"player_changename":               handleGenericEvent,               // Name change
	"buytime_ended":                   handleGenericEvent,               // Not actually end of buy time, seems to only be sent once per game at the start
	"round_announce_match_start":      handleGenericEvent,               // Special match start announcement
	"bomb_beep":                       handleGenericEvent,               // Bomb beep
	"player_spawn":                    handleGenericEvent,               // Player spawn
	"hltv_status":                     handleGenericEvent,               // Don't know
	"hltv_chase":                      handleGenericEvent,               // Don't care
	"cs_round_start_beep":             handleGenericEvent,               // Round start beeps
	"cs_round_final_beep":             handleGenericEvent,               // Final beep
	"cs_pre_restart":                  handleGenericEvent,               // Not sure, doesn't seem to be important
	"round_prestart":                  handleGenericEvent,               // Ditto
	"round_poststart":                 handleGenericEvent,               // Ditto
	"cs_win_panel_round":              handleGenericEvent,               // Win panel, (==end of match?)
	"endmatch_cmm_start_reveal_items": handleGenericEvent,               // Drops
	"announce_phase_end":              handleGenericEvent,               // Dunno
	"tournament_reward":               handleGenericEvent,               // Dunno
	"other_death":                     handleGenericEvent,               // Dunno
	"round_announce_warmup":           handleGenericEvent,               // Dunno
	"server_cvar":                     handleGenericEvent,               // Dunno
	"weapon_fire_on_empty":            handleGenericEvent,               // Sounds boring
	"hltv_fixed":                      handleGenericEvent,               // Dunno
	"cs_match_end_restart":            handleGenericEvent,               // Yawn
}

// TODO: remove return shit
type gameEventHandler func(*Parser, *msg.CSVCMsg_GameEventListDescriptorT, *msg.CSVCMsg_GameEvent)

func handleRoundStart(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)
	p.eventDispatcher.Dispatch(events.RoundStart{
		TimeLimit: int(data["timelimit"].GetValLong()),
		FragLimit: int(data["fraglimit"].GetValLong()),
		Objective: data["objective"].GetValString(),
	})
}

func handleCsWinPanelMatch(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.AnnouncementWinPanelMatch{})
}

func handleRoundAnnounceFinal(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.AnnouncementFinalRound{})
}

func handleRoundAnnounceLastRoundHalf(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.AnnouncementLastRoundHalf{})
}

func handleRoundEnd(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	winner := common.Team(data["winner"].ValByte)
	winnerState := p.gameState.Team(winner)
	var loserState *common.TeamState
	if winnerState != nil {
		loserState = winnerState.Opponent
	}

	p.eventDispatcher.Dispatch(events.RoundEnd{
		Message:     data["message"].GetValString(),
		Reason:      events.RoundEndReason(data["reason"].GetValByte()),
		Winner:      winner,
		WinnerState: winnerState,
		LoserState:  loserState,
	})
}

func handleRoundOfficiallyEnded(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	// Issue #42
	// Sometimes grenades & infernos aren't deleted / destroyed via entity-updates at the end of the round,
	// so we need to do it here for those that weren't.
	//
	// We're not deleting them from entitites though as that's supposed to be as close to the actual demo data as possible.
	// We're also not using Entity.Destroy() because it would - in some cases - be called twice on the same entity
	// and it's supposed to be called when the demo actually says so (same case as with GameState.entities).
	for _, proj := range p.gameState.grenadeProjectiles {
		p.nadeProjectileDestroyed(proj)
	}

	for _, inf := range p.gameState.infernos {
		p.infernoExpired(inf)
	}

	p.eventDispatcher.Dispatch(events.RoundEndOfficial{})
}

func handleRoundMVP(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	p.eventDispatcher.Dispatch(events.RoundMVPAnnouncement{
		Player: p.gameState.playersByUserID[int(data["userid"].GetValShort())],
		Reason: events.RoundMVPReason(data["reason"].GetValShort()),
	})
}

func handleBotTakeover(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	p.eventDispatcher.Dispatch(events.BotTakenOver{
		Taker: p.gameState.playersByUserID[int(data["userid"].GetValShort())],
	})
}

func handleBeginNewMatch(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.MatchStart{})
}

func handleRoundFreezeEnd(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.RoundFreezetimeEnd{})
}

func handlePlayerFootstep(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	p.eventDispatcher.Dispatch(events.Footstep{
		Player: p.gameState.playersByUserID[int(data["userid"].GetValShort())],
	})
}

func handlePlayerJump(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	p.eventDispatcher.Dispatch(events.PlayerJump{
		Player: p.gameState.playersByUserID[int(data["userid"].GetValShort())],
	})
}

func handleWeaponFire(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	shooter := p.gameState.playersByUserID[int(data["userid"].GetValShort())]
	wepType := common.MapEquipment(data["weapon"].GetValString())

	p.eventDispatcher.Dispatch(events.WeaponFire{
		Shooter: shooter,
		Weapon:  getPlayerWeapon(shooter, wepType),
	})
}

func handlePlayerDeath(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	killer := p.gameState.playersByUserID[int(data["attacker"].GetValShort())]
	wepType := common.MapEquipment(data["weapon"].GetValString())

	p.eventDispatcher.Dispatch(events.Kill{
		Victim:            p.gameState.playersByUserID[int(data["userid"].GetValShort())],
		Killer:            killer,
		Assister:          p.gameState.playersByUserID[int(data["assister"].GetValShort())],
		IsHeadshot:        data["headshot"].GetValBool(),
		PenetratedObjects: int(data["penetrated"].GetValShort()),
		Weapon:            getPlayerWeapon(killer, wepType),
	})
}

func handlePlayerHurt(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	attacker := p.gameState.playersByUserID[int(data["attacker"].GetValShort())]
	wepType := common.MapEquipment(data["weapon"].GetValString())

	p.eventDispatcher.Dispatch(events.PlayerHurt{
		Player:       p.gameState.playersByUserID[int(data["userid"].GetValShort())],
		Attacker:     attacker,
		Health:       int(data["health"].GetValByte()),
		Armor:        int(data["armor"].GetValByte()),
		HealthDamage: int(data["dmg_health"].GetValShort()),
		ArmorDamage:  int(data["dmg_armor"].GetValByte()),
		HitGroup:     events.HitGroup(data["hitgroup"].GetValByte()),
		Weapon:       getPlayerWeapon(attacker, wepType),
	})
}

func handlePlayerBlind(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	// Player.FlashDuration hasn't been updated yet,
	// so we need to wait until the end of the tick before dispatching
	p.eventDispatcher.Dispatch(events.PlayerFlashed{
		Player:   p.gameState.playersByUserID[int(data["userid"].GetValShort())],
		Attacker: p.gameState.lastFlasher,
	})
}

func handleFlashbangDetonate(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	nadeEvent := handleNadeEvent(p, desc, ge, common.EqFlash)

	p.gameState.lastFlasher = nadeEvent.Thrower
	p.eventDispatcher.Dispatch(events.FlashExplode{
		GrenadeEvent: nadeEvent,
	})
}

func handleHegranadeDetonate(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.HeExplode{
		GrenadeEvent: handleNadeEvent(p, desc, ge, common.EqHE),
	})
}

func handleDecoyStarted(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.DecoyStart{
		GrenadeEvent: handleNadeEvent(p, desc, ge, common.EqDecoy),
	})
}

func handleDecoyDetonate(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.DecoyExpired{
		GrenadeEvent: handleNadeEvent(p, desc, ge, common.EqDecoy),
	})
}

func handleSmokegrenadeDetonate(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.SmokeStart{
		GrenadeEvent: handleNadeEvent(p, desc, ge, common.EqSmoke),
	})
}

func handleSmokegrenadeExpired(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.SmokeExpired{
		GrenadeEvent: handleNadeEvent(p, desc, ge, common.EqSmoke),
	})
}

func handleInfernoStartburn(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.FireGrenadeStart{
		GrenadeEvent: handleNadeEvent(p, desc, ge, common.EqIncendiary),
	})
}

func handleInfernoExpire(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.FireGrenadeExpired{
		GrenadeEvent: handleNadeEvent(p, desc, ge, common.EqIncendiary),
	})
}

func handlePlayerConnect(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	pl := &playerInfo{
		userID: int(data["userid"].GetValShort()),
		name:   data["name"].GetValString(),
		guid:   data["networkid"].GetValString(),
	}

	pl.xuid = getCommunityID(pl.guid)

	p.rawPlayers[int(data["index"].GetValByte())] = pl
}

func handlePlayerDisconnect(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	uid := int(data["userid"].GetValShort())

	for k, v := range p.rawPlayers {
		if v.userID == uid {
			delete(p.rawPlayers, k)
		}
	}

	pl := p.gameState.playersByUserID[uid]
	if pl != nil {
		// Dispatch this event early since we delete the player on the next line
		p.eventDispatcher.Dispatch(events.PlayerDisconnected{
			Player: pl,
		})
	}

	delete(p.gameState.playersByUserID, uid)
}

func handlePlayerTeam(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	player := p.gameState.playersByUserID[int(data["userid"].GetValShort())]
	newTeam := common.Team(data["team"].GetValByte())

	if player != nil {
		if player.Team != newTeam {
			player.Team = newTeam

			oldTeam := common.Team(data["oldteam"].GetValByte())
			p.eventDispatcher.Dispatch(events.PlayerTeamChange{
				Player:       player,
				IsBot:        data["isbot"].GetValBool(),
				Silent:       data["silent"].GetValBool(),
				NewTeam:      newTeam,
				NewTeamState: p.gameState.Team(newTeam),
				OldTeam:      oldTeam,
				OldTeamState: p.gameState.Team(oldTeam),
			})
		} else {
			p.eventDispatcher.Dispatch(events.ParserWarn{
				Message: "Player team swap game-event occurred but player.Team == newTeam",
			})
		}
	} else {
		p.eventDispatcher.Dispatch(events.ParserWarn{
			Message: "Player team swap game-event occurred but player is nil",
		})
	}
}

func handleBombBeginplant(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.BombPlantBegin{BombEvent: handleBombEvent(p, desc, ge)})
}

func handleBombPlanted(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.BombPlanted{BombEvent: handleBombEvent(p, desc, ge)})
}

func handleBombDefused(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	bombEvent := handleBombEvent(p, desc, ge)
	p.gameState.currentDefuser = nil
	p.eventDispatcher.Dispatch(events.BombDefused{BombEvent: bombEvent})
}

func handleBombExploded(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	bombEvent := handleBombEvent(p, desc, ge)
	p.gameState.currentDefuser = nil
	p.eventDispatcher.Dispatch(events.BombExplode{BombEvent: bombEvent})
}

func handleBombEvent(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) events.BombEvent {
	data := mapGameEventData(desc, ge)

	bombEvent := events.BombEvent{Player: p.gameState.playersByUserID[int(data["userid"].GetValShort())]}

	site := int(data["site"].GetValShort())

	switch site {
	case p.bombsiteA.index:
		bombEvent.Site = events.BombsiteA
	case p.bombsiteB.index:
		bombEvent.Site = events.BombsiteB
	default:
		t := p.triggers[site]

		if t == nil {
			panic(fmt.Sprintf("Bombsite with index %d not found", site))
		}

		if t.contains(p.bombsiteA.center) {
			bombEvent.Site = events.BombsiteA
			p.bombsiteA.index = site
		} else if t.contains(p.bombsiteB.center) {
			bombEvent.Site = events.BombsiteB
			p.bombsiteB.index = site
		} else {
			panic("Bomb not planted on bombsite A or B")
		}
	}

	return bombEvent
}

func handleBombBegindefuse(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	p.gameState.currentDefuser = p.gameState.playersByUserID[int(data["userid"].GetValShort())]

	p.eventDispatcher.Dispatch(events.BombDefuseStart{
		Player: p.gameState.currentDefuser,
		HasKit: data["haskit"].GetValBool(),
	})
}

func handleItemEquip(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	player, weapon := handleItemEvent(p, desc, ge)
	p.eventDispatcher.Dispatch(events.ItemEquip{
		Player: player,
		Weapon: weapon,
	})
}

func handleItemPickup(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	player, weapon := handleItemEvent(p, desc, ge)
	p.eventDispatcher.Dispatch(events.ItemPickup{
		Player: player,
		Weapon: weapon,
	})
}

func handleItemRemove(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	player, weapon := handleItemEvent(p, desc, ge)
	p.eventDispatcher.Dispatch(events.ItemDrop{
		Player: player,
		Weapon: weapon,
	})
}

func handleItemEvent(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) (*common.Player, common.Equipment) {
	data := mapGameEventData(desc, ge)
	player := p.gameState.playersByUserID[int(data["userid"].GetValShort())]

	wepType := common.MapEquipment(data["item"].GetValString())
	weapon := common.NewEquipment(wepType)

	return player, weapon
}

func handleBombDropped(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	player := p.gameState.playersByUserID[int(data["userid"].GetValShort())]
	entityID := int(data["entityid"].GetValShort())

	p.eventDispatcher.Dispatch(events.BombDropped{
		Player:   player,
		EntityID: entityID,
	})
}

func handleBombPickup(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	data := mapGameEventData(desc, ge)

	p.eventDispatcher.Dispatch(events.BombPickup{
		Player: p.gameState.playersByUserID[int(data["userid"].GetValShort())],
	})
}

func handleGenericEvent(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.GenericGameEvent{
		Name: desc.Name,
		Data: mapGameEventData(desc, ge),
	})
}

func handleUnknownEvent(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent) {
	p.eventDispatcher.Dispatch(events.ParserWarn{Message: fmt.Sprintf("Unknown event %q", desc.Name)})
	handleGenericEvent(p, desc, ge)
}

// Returns the players instance of the weapon if applicable or a new instance otherwise.
func getPlayerWeapon(player *common.Player, wepType common.EquipmentElement) *common.Equipment {
	class := wepType.Class()
	isSpecialWeapon := class == common.EqClassGrenade || (class == common.EqClassEquipment && wepType != common.EqKnife)
	if !isSpecialWeapon && player != nil {
		for _, wep := range player.Weapons() {
			if wep.Weapon == wepType {
				return wep
			}
		}
	}

	wep := common.NewEquipment(wepType)
	return &wep
}

func mapGameEventData(d *msg.CSVCMsg_GameEventListDescriptorT, e *msg.CSVCMsg_GameEvent) map[string]*msg.CSVCMsg_GameEventKeyT {
	data := make(map[string]*msg.CSVCMsg_GameEventKeyT)
	for i, k := range d.Keys {
		data[k.Name] = e.Keys[i]
	}
	return data
}

// Just so we can nicely create GrenadeEvents in one line
func handleNadeEvent(p *Parser, desc *msg.CSVCMsg_GameEventListDescriptorT, ge *msg.CSVCMsg_GameEvent, nadeType common.EquipmentElement) events.GrenadeEvent {
	data := mapGameEventData(desc, ge)
	thrower := p.gameState.playersByUserID[int(data["userid"].GetValShort())]
	position := r3.Vector{
		X: float64(data["x"].ValFloat),
		Y: float64(data["y"].ValFloat),
		Z: float64(data["z"].ValFloat),
	}
	nadeEntityID := int(data["entityid"].GetValShort())

	return events.GrenadeEvent{
		GrenadeType:     nadeType,
		Thrower:         thrower,
		Position:        position,
		GrenadeEntityID: nadeEntityID,
	}
}

// We're all better off not asking questions
const valveMagicNumber = 76561197960265728

func getCommunityID(guid string) int64 {
	if guid == "BOT" {
		return 0
	}

	authSrv, errSrv := strconv.ParseInt(guid[8:9], 10, 64)
	authID, errID := strconv.ParseInt(guid[10:], 10, 64)

	if errSrv != nil {
		panic(errSrv.Error())
	}
	if errID != nil {
		panic(errID.Error())
	}

	// WTF are we doing here?
	return valveMagicNumber + authID*2 + authSrv
}

func (p *Parser) handleUserMessage(um *msg.CSVCMsg_UserMessage) {
	defer func() {
		p.setError(recoverFromUnexpectedEOF(recover()))
	}()

	switch msg.ECstrike15UserMessages(um.MsgType) {
	case msg.ECstrike15UserMessages_CS_UM_SayText:
		st := new(msg.CCSUsrMsg_SayText)
		err := st.Unmarshal(um.MsgData)
		if err != nil {
			p.eventDispatcher.Dispatch(events.ParserWarn{Message: fmt.Sprintf("Failed to decode SayText message: %s", err.Error())})
		}

		p.eventDispatcher.Dispatch(events.SayText{
			EntIdx:    int(st.EntIdx),
			IsChat:    st.Chat,
			IsChatAll: st.Textallchat,
			Text:      st.Text,
		})

	case msg.ECstrike15UserMessages_CS_UM_SayText2:
		st := new(msg.CCSUsrMsg_SayText2)
		err := st.Unmarshal(um.MsgData)
		if err != nil {
			p.eventDispatcher.Dispatch(events.ParserWarn{Message: fmt.Sprintf("Failed to decode SayText2 message: %s", err.Error())})
		}

		p.eventDispatcher.Dispatch(events.SayText2{
			EntIdx:    int(st.EntIdx),
			IsChat:    st.Chat,
			IsChatAll: st.Textallchat,
			MsgName:   st.MsgName,
			Params:    st.Params,
		})

		switch st.MsgName {
		case "Cstrike_Chat_All":
			fallthrough
		case "Cstrike_Chat_AllDead":
			var sender *common.Player
			for _, pl := range p.gameState.playersByUserID {
				// This could be a problem if the player changed his name
				// as the name is only set initially and never updated
				if pl.Name == st.Params[0] {
					sender = pl
				}
			}

			p.eventDispatcher.Dispatch(events.ChatMessage{
				Sender:    sender,
				Text:      st.Params[1],
				IsChatAll: st.Textallchat,
			})

		case "#CSGO_Coach_Join_T": // Ignore these
		case "#CSGO_Coach_Join_CT":

		default:
			p.eventDispatcher.Dispatch(events.ParserWarn{Message: fmt.Sprintf("Skipped sending ChatMessageEvent for SayText2 with unknown MsgName %q", st.MsgName)})
		}

	default:
		// TODO: handle more user messages (if they are interesting)
		// Maybe msg.ECstrike15UserMessages_CS_UM_RadioText
	}
}
