package game_net

import (
	"hmc_server.com/hmc_server/src/database"
	"hmc_server.com/hmc_server/src/transform"
)

type FRecvPacket_PlayerLogout struct {
	PacketName string
	Id         string
}

func New_FRecvPacket_PlayerLogout() FRecvPacket_PlayerLogout {
	result := FRecvPacket_PlayerLogout{}
	result.PacketName = "FRecvPacket_PlayerLogout"
	return result
}

type FRecvPacket_OtherPlayerMove struct {
	PacketName   string
	Id           string
	Destination  transform.Vector3
	DestRotation transform.Vector3
	MoveSpeed    float32
	RotateSpeed  float32
}

func New_FRecvPacket_OtherPlayerMove() FRecvPacket_OtherPlayerMove {
	result := FRecvPacket_OtherPlayerMove{}
	result.PacketName = "FRecvPacket_OtherPlayerMove"
	return result
}

type FRecvPacket_OtherPlayerSpawnInfo struct {
	PacketName string
	Id         string
	Position   transform.Vector3
	Rotation   transform.Vector3

	IsMan          bool
	SkinIndex      int32
	TopIndex       int32
	BottomIndex    int32
	HairIndex      int32
	ShoesIndex     int32
	HairColorIndex int32
	FaceIndex      int32
	AccessoryIndex int32

	FirstName  string
	LastName   string
	BirthDay   int32
	BirthMonth int32
	DealerType string
	Country    string

	IsWearableWear bool
	IsDancing      bool
	IsSitting      bool
	StatusNum      int32

	IsAtlasSpawn  bool
	AtlasPosition transform.Vector3
	AtlasRotation transform.Vector3

	IsSpotSpawn  bool
	SpotPosition transform.Vector3
	SpotRotation transform.Vector3
}

func New_FRecvPacket_OtherPlayerSpawnInfo() FRecvPacket_OtherPlayerSpawnInfo {
	result := FRecvPacket_OtherPlayerSpawnInfo{}
	result.PacketName = "FRecvPacket_OtherPlayerSpawnInfo"
	return result
}

type FRecvPacket_OtherPlayerDestroyInfo struct {
	PacketName string
	Id         string
}

func New_FRecvPacket_OtherPlayerDestroyInfo() FRecvPacket_OtherPlayerDestroyInfo {
	result := FRecvPacket_OtherPlayerDestroyInfo{}
	result.PacketName = "FRecvPacket_OtherPlayerDestroyInfo"
	return result
}

type FRecvPacket_LoginInfo struct {
	PacketName string
	List       []FRecvPacket_OtherPlayerSpawnInfo
	NPCList    []FRecvPacket_NPCSpawnInfo
	Position   transform.Vector3
}

func New_FRecvPacket_LoginInfo() FRecvPacket_LoginInfo {
	result := FRecvPacket_LoginInfo{}
	result.PacketName = "FRecvPacket_LoginInfo"
	return result
}

type FRecvPacket_GetModerator struct {
	PacketName string
	Moderators []database.Moderator
}

func New_FRecvPacket_GetModerator() FRecvPacket_GetModerator {
	result := FRecvPacket_GetModerator{}
	result.PacketName = "FRecvPacket_GetModerator"
	return result
}

type ModeratorUserInfo struct {
	Id          string
	URL         string
	IsModerator bool
	UserName    string
	OnMic       bool
	OnHandsUp   bool
}

type FRecvPacket_JoinVoiceGroupUpdate struct {
	PacketName string
	Ids        []ModeratorUserInfo
}

func New_FRecvPacket_JoinVoiceGroupUpdate() FRecvPacket_JoinVoiceGroupUpdate {
	result := FRecvPacket_JoinVoiceGroupUpdate{}
	result.PacketName = "FRecvPacket_JoinVoiceGroupUpdate"
	return result
}

type FRecvPacket_LeaveVoiceGroupUpdate struct {
	PacketName string
	Ids        []ModeratorUserInfo
}

func New_FRecvPacket_LeaveVoiceGroupUpdate() FRecvPacket_LeaveVoiceGroupUpdate {
	result := FRecvPacket_LeaveVoiceGroupUpdate{}
	result.PacketName = "FRecvPacket_LeaveVoiceGroupUpdate"
	return result
}

type FRecvPacket_SendToVoiceChat struct {
	PacketName string
	Id         string
	UserName   string
	Message    string
}

func New_FRecvPacket_SendToVoiceChat() FRecvPacket_SendToVoiceChat {
	result := FRecvPacket_SendToVoiceChat{}
	result.PacketName = "FRecvPacket_SendToVoiceChat"
	return result
}

type FRecvPacket_VoiceHandUp struct {
	PacketName string
	HandUpId   string
}

func New_FRecvPacket_VoiceHandUp() FRecvPacket_VoiceHandUp {
	result := FRecvPacket_VoiceHandUp{}
	result.PacketName = "FRecvPacket_VoiceHandUp"
	return result
}

type FRecvPacket_VoicePlayerChoice struct {
	PacketName string
	Id         string
	ChoiceUpId string
	IsCancel   bool
}

func New_FRecvPacket_VoicePlayerChoice() FRecvPacket_VoicePlayerChoice {
	result := FRecvPacket_VoicePlayerChoice{}
	result.PacketName = "FRecvPacket_VoicePlayerChoice"
	return result
}

type FRecvPacket_VoicePlayerChoiceCancle struct {
	PacketName string
	Id         string
	CancleId   string
}

func New_FRecvPacket_VoicePlayerChoiceCancle() FRecvPacket_VoicePlayerChoiceCancle {
	result := FRecvPacket_VoicePlayerChoiceCancle{}
	result.PacketName = "FRecvPacket_VoicePlayerChoiceCancle"
	return result
}

type FRecvPacket_NPCMove struct {
	PacketName  string
	Id          string
	Destination transform.Vector3
	MoveSpeed   float32
}

func New_FRecvPacket_NPCMove() FRecvPacket_NPCMove {
	result := FRecvPacket_NPCMove{}
	result.PacketName = "FRecvPacket_NPCMove"
	return result
}

type FRecvPacket_NPCSpawnInfo struct {
	PacketName string
	Id         string
	Position   transform.Vector3
}

func New_FRecvPacket_NPCSpawnInfo() FRecvPacket_NPCSpawnInfo {
	result := FRecvPacket_NPCSpawnInfo{}
	result.PacketName = "FRecvPacket_NPCSpawnInfo"
	return result
}

type FRecvPacket_NearPlayerUpdate struct {
	PacketName  string
	SpawnList   []FRecvPacket_OtherPlayerSpawnInfo
	DestroyList []FRecvPacket_OtherPlayerDestroyInfo
}

func New_FRecvPacket_NearPlayerUpdate() FRecvPacket_NearPlayerUpdate {
	result := FRecvPacket_NearPlayerUpdate{}
	result.PacketName = "FRecvPacket_NearPlayerUpdate"
	return result
}

type FRecvPacket_PlayerActionEvent struct {
	PacketName string
	Id         string
	ActionId   string
}

func New_FRecvPacket_PlayerActionEvent() FRecvPacket_PlayerActionEvent {
	result := FRecvPacket_PlayerActionEvent{}
	result.PacketName = "FRecvPacket_PlayerActionEvent"
	return result
}

type FRecvPacket_Voice struct {
	PacketName  string
	Id          string
	VoiceData   []uint16
	Numchannels int32
	SampleRate  int32
	PCMSize     int32
}

func New_FRecvPacket_Voice() FRecvPacket_Voice {
	result := FRecvPacket_Voice{}
	result.PacketName = "FRecvPacket_Voice"
	return result
}

type FRecvPacket_GlobalChat struct {
	PacketName string
	Id         string
	UserName   string
	Message    string
}

func New_FRecvPacket_GlobalChat() FRecvPacket_GlobalChat {
	result := FRecvPacket_GlobalChat{}
	result.PacketName = "FRecvPacket_GlobalChat"
	return result
}

type FRecvPacket_PrivateChat struct {
	PacketName string
	Id         string
	TargetId   string
	UserName   string
	Message    string
	IsOnline   bool
}

func New_FRecvPacket_PrivateChat() FRecvPacket_PrivateChat {
	result := FRecvPacket_PrivateChat{}
	result.PacketName = "FRecvPacket_PrivateChat"
	return result
}

type FRecvPacket_Notice struct {
	PacketName string
	Message    string
}

func New_FRecvPacket_Notice() FRecvPacket_Notice {
	result := FRecvPacket_Notice{}
	result.PacketName = "FRecvPacket_Notice"
	return result
}

type FRecvPacket_CreateTreasureBox struct {
	PacketName string
	Id         string
	Point      int32
	Position   transform.Vector3
	LevelName  string
}

func New_FRecvPacket_CreateTreasureBox() FRecvPacket_CreateTreasureBox {
	result := FRecvPacket_CreateTreasureBox{}
	result.PacketName = "FRecvPacket_CreateTreasureBox"
	return result
}

type FRecvPacket_DestroyTreasureBox struct {
	PacketName string
	Id         string
}

func New_FRecvPacket_DestroyTreasureBox() FRecvPacket_DestroyTreasureBox {
	result := FRecvPacket_DestroyTreasureBox{}
	result.PacketName = "FRecvPacket_DestroyTreasureBox"
	return result
}

type FRecvPacket_TreasureBoxInfo struct {
	PacketName  string
	BoxList     []FRecvPacket_CreateTreasureBox
	DestroyList []FRecvPacket_DestroyTreasureBox
}

func New_FRecvPacket_TreasureBoxInfo() FRecvPacket_TreasureBoxInfo {
	result := FRecvPacket_TreasureBoxInfo{}
	result.PacketName = "FRecvPacket_TreasureBoxInfo"
	return result
}

type FRecvPacket_WorldTeleport struct {
	PacketName string
	Id         string
	Location   transform.Vector3
}

func New_FRecvPacket_WorldTeleport() FRecvPacket_WorldTeleport {
	result := FRecvPacket_WorldTeleport{}
	result.PacketName = "FRecvPacket_WorldTeleport"
	return result
}

type FRecvPacket_OtherCarMove struct {
	PacketName     string
	Id             string
	ServerDistacne float64
}

func New_FRecvPacket_OtherCarMove() FRecvPacket_OtherCarMove {
	result := FRecvPacket_OtherCarMove{}
	result.PacketName = "FRecvPacket_OtherCarMove"
	return result
}

type FRecvPacket_OtherCarSpawnInfo struct {
	PacketName     string
	Id             string
	TypeNum        int32
	PathTag        string
	ServerDistacne float64
}

func New_FRecvPacket_OtherCarSpawnInfo() FRecvPacket_OtherCarSpawnInfo {
	result := FRecvPacket_OtherCarSpawnInfo{}
	result.PacketName = "FRecvPacket_OtherCarSpawnInfo"
	result.ServerDistacne = 0
	return result
}

type FRecvPacket_OtherCarDestroy struct {
	PacketName string
	Id         string
}

func New_FRecvPacket_OtherCarDestroy() FRecvPacket_OtherCarDestroy {
	result := FRecvPacket_OtherCarDestroy{}
	result.PacketName = "FRecvPacket_OtherCarDestroy"
	return result
}

type FRecvPacket_Wearable struct {
	PacketName string
	Id         string
	IsWear     bool
}

func New_FRecvPacket_Wearable() FRecvPacket_Wearable {
	result := FRecvPacket_Wearable{}
	result.PacketName = "FRecvPacket_Wearable"
	return result
}

type FRecvPacket_RHQPBV_Move_Timeline struct {
	PacketName  string
	MoveForward float32
}

func New_FRecvPacket_RHQPBV_Move_Timeline() FRecvPacket_RHQPBV_Move_Timeline {
	result := FRecvPacket_RHQPBV_Move_Timeline{}
	result.PacketName = "FRecvPacket_RHQPBV_Move_Timeline"
	return result
}

type FRecvPacket_SpotSpawn struct {
	PacketName string
	Id         string
	SpotId     string
	SpawnPoint transform.Vector3
}

func New_FRecvPacket_SpotSpawn() FRecvPacket_SpotSpawn {
	result := FRecvPacket_SpotSpawn{}
	result.PacketName = "FRecvPacket_SpotSpawn"
	return result
}

type FRecvPacket_SpotMove struct {
	PacketName   string
	Id           string
	Destination  transform.Vector3
	DestRotation transform.Vector3
	MoveSpeed    float32
	RotateSpeed  float32
}

func New_FRecvPacket_SpotMove() FRecvPacket_SpotMove {
	result := FRecvPacket_SpotMove{}
	result.PacketName = "FRecvPacket_SpotMove"
	return result
}

type FRecvPacket_SpotDestroy struct {
	PacketName string
	Id         string
	SpotId     string
}

func New_FRecvPacket_SpotDestroy() FRecvPacket_SpotDestroy {
	result := FRecvPacket_SpotDestroy{}
	result.PacketName = "FRecvPacket_SpotDestroy"
	return result
}

type FRecvPacket_AtlasSpawn struct {
	PacketName string
	Id         string
	SpawnPoint transform.Vector3
}

func New_FRecvPacket_AtlasSpawn() FRecvPacket_AtlasSpawn {
	result := FRecvPacket_AtlasSpawn{}
	result.PacketName = "FRecvPacket_AtlasSpawn"
	return result
}

type FRecvPacket_AtlasMove struct {
	PacketName   string
	Id           string
	Destination  transform.Vector3
	DestRotation transform.Vector3
	MoveSpeed    float32
	RotateSpeed  float32
}

func New_FRecvPacket_AtlasMove() FRecvPacket_AtlasMove {
	result := FRecvPacket_AtlasMove{}
	result.PacketName = "FRecvPacket_AtlasMove"
	return result
}

type FRecvPacket_AtlasDestroy struct {
	PacketName string
	Id         string
	AtlasId    string
}

func New_FRecvPacket_AtlasDestroy() FRecvPacket_AtlasDestroy {
	result := FRecvPacket_AtlasDestroy{}
	result.PacketName = "FRecvPacket_AtlasDestroy"
	return result
}

type FRecvPacket_SpotInfo struct {
	PacketName  string
	SpawnList   []FRecvPacket_SpotSpawn
	DestroyList []FRecvPacket_SpotDestroy
}

func New_FRecvPacket_SpotInfo() FRecvPacket_SpotInfo {
	result := FRecvPacket_SpotInfo{}
	result.PacketName = "FRecvPacket_SpotInfo"
	return result
}

type FRecvPacket_AtlasInfo struct {
	PacketName  string
	SpawnList   []FRecvPacket_AtlasSpawn
	DestroyList []FRecvPacket_AtlasDestroy
}

func New_FRecvPacket_AtlasInfo() FRecvPacket_AtlasInfo {
	result := FRecvPacket_AtlasInfo{}
	result.PacketName = "FRecvPacket_AtlasInfo"
	return result
}

type FRecvPacket_CarInfos struct {
	PacketName string
	List       []FRecvPacket_OtherCarSpawnInfo
}

func New_FRecvPacket_CarInfos() FRecvPacket_CarInfos {
	result := FRecvPacket_CarInfos{}
	result.PacketName = "FRecvPacket_CarInfos"
	return result
}

type FRecvPacket_CreateChatGroup struct {
	PacketName string
	Id         string
	IsCreate   bool
}

func New_FRecvPacket_CreateChatGroup() FRecvPacket_CreateChatGroup {
	result := FRecvPacket_CreateChatGroup{}
	result.PacketName = "FRecvPacket_CreateChatGroup"
	return result
}

type FRecvPacket_JoinChatGroup struct {
	PacketName string
	Id         string
	IsJoined   bool
}

func New_FRecvPacket_JoinChatGroup() FRecvPacket_JoinChatGroup {
	result := FRecvPacket_JoinChatGroup{}
	result.PacketName = "FRecvPacket_JoinChatGroup"
	return result
}

type FRecvPacket_LeaveChatGroup struct {
	PacketName string
	Id         string
	IsLeaved   bool
}

func New_FRecvPacket_LeaveChatGroup() FRecvPacket_LeaveChatGroup {
	result := FRecvPacket_LeaveChatGroup{}
	result.PacketName = "FRecvPacket_LeaveChatGroup"
	return result
}

type FRecvPacket_SendToChatGroup struct {
	PacketName string
	Id         string
	Message    string
}

func New_FRecvPacket_SendToChatGroup() FRecvPacket_SendToChatGroup {
	result := FRecvPacket_SendToChatGroup{}
	result.PacketName = "FRecvPacket_SendToChatGroup"
	return result
}

type FRecvPacket_GroupInvitation struct {
	PacketName string
	Id         string
	TargetId   string
	GroupId    string
}

func New_FRecvPacket_GroupInvitation() FRecvPacket_GroupInvitation {
	result := FRecvPacket_GroupInvitation{}
	result.PacketName = "FRecvPacket_GroupInvitation"
	return result
}

type FRecvPacket_AcceptGroupInvitation struct {
	PacketName string
	Id         string
	GroupId    string
}

func New_FRecvPacket_AcceptGroupInvitation() FRecvPacket_AcceptGroupInvitation {
	result := FRecvPacket_AcceptGroupInvitation{}
	result.PacketName = "FRecvPacket_AcceptGroupInvitation"
	return result
}

type Group struct {
	GroupId    string
	IsPassword bool
}

type FRecvPacket_GroupList struct {
	PacketName string
	GroupList  []Group
}

func New_FRecvPacket_GroupList() FRecvPacket_GroupList {
	result := FRecvPacket_GroupList{}
	result.PacketName = "FRecvPacket_GroupList"
	return result
}

type FRecvPacket_SetCostume struct {
	PacketName string
	Id         string
	IsMan      bool
	ClothType  int32
	ClothIndex int32
}

func New_FRecvPacket_SetCostume() FRecvPacket_SetCostume {
	result := FRecvPacket_SetCostume{}
	result.PacketName = "FRecvPacket_SetCostume"
	return result
}

type WRCRankInfo struct {
	UserName string
	Laptime  int
}

type FRecvPacket_WRCRankUpdate struct {
	PacketName string
	Ranking    []WRCRankInfo
}

func New_FRecvPacket_WRCRankUpdate() FRecvPacket_WRCRankUpdate {
	result := FRecvPacket_WRCRankUpdate{}
	result.PacketName = "FRecvPacket_WRCRankUpdate"
	return result
}

type FRecvPacket_HandUp struct {
	PacketName string
	Id         string
	HandUp     bool
}

func New_FRecvPacket_HandUp() FRecvPacket_HandUp {
	result := FRecvPacket_HandUp{}
	result.PacketName = "FRecvPacket_HandUp"
	return result
}

type FRecvPacket_MicOn struct {
	PacketName string
	Id         string
}

func New_FRecvPacket_MicOn() FRecvPacket_MicOn {
	result := FRecvPacket_MicOn{}
	result.PacketName = "FRecvPacket_MicOn"
	return result
}

type FRecvPacket_MicOff struct {
	PacketName string
	Id         string
}

func New_FRecvPacket_MicOff() FRecvPacket_MicOff {
	result := FRecvPacket_MicOff{}
	result.PacketName = "FRecvPacket_MicOff"
	return result
}

type FRecvPacket_CharacterStatus struct {
	PacketName string
	Id         string
	Status     int32
}

func New_FRecvPacket_CharacterStatus() FRecvPacket_CharacterStatus {
	result := FRecvPacket_CharacterStatus{}
	result.PacketName = "FRecvPacket_CharacterStatus"
	return result
}

type FRecvPacket_StartLuckyDraw struct {
	PacketName   string
	LuckyNumbers []string
}

func New_FRecvPacket_StartLuckyDraw() FRecvPacket_StartLuckyDraw {
	result := FRecvPacket_StartLuckyDraw{}
	result.PacketName = "FRecvPacket_StartLuckyDraw"
	return result
}

type FRecvPacket_LuckyDrawWinner struct {
	PacketName string
	Reward     int32
}

func New_FRecvPacket_LuckyDrawWinner() FRecvPacket_LuckyDrawWinner {
	result := FRecvPacket_LuckyDrawWinner{}
	result.PacketName = "FRecvPacket_LuckyDrawWinner"
	return result
}

type FRecvPacket_GMSequencePlay struct {
	PacketName   string
	TotalSeconds int32
	SequenceNum  int32
}

func New_FRecvPacket_GMSequencePlay() FRecvPacket_GMSequencePlay {
	result := FRecvPacket_GMSequencePlay{}
	result.PacketName = "FRecvPacket_GMSequencePlay"
	return result
}

type FRecvPacket_SequenceSyncNewPlayer struct {
	PacketName string
	Id         string
}

func New_FRecvPacket_SequenceSyncNewPlayer() FRecvPacket_SequenceSyncNewPlayer {
	result := FRecvPacket_SequenceSyncNewPlayer{}
	result.PacketName = "FRecvPacket_SequenceSyncNewPlayer"
	return result
}

type FRecvPacket_FindPassword struct {
	PacketName string
	Status     bool
}

func New_FRecvPacket_FindPassword() FRecvPacket_FindPassword {
	result := FRecvPacket_FindPassword{}
	result.PacketName = "FRecvPacket_FindPassword"
	return result
}

type FRecvPacket_VerifyCode struct {
	PacketName string
	Status     bool
}

func New_FRecvPacket_VerifyCode() FRecvPacket_VerifyCode {
	result := FRecvPacket_VerifyCode{}
	result.PacketName = "FRecvPacket_VerifyCode"
	return result
}

type FRecvPacket_ResetPassword struct {
	PacketName string
	Status     bool
}

func New_FRecvPacket_ResetPassword() FRecvPacket_ResetPassword {
	result := FRecvPacket_ResetPassword{}
	result.PacketName = "FRecvPacket_ResetPassword"
	return result
}

type FRecvPacket_MTHText struct {
	PacketName  string
	LeftMTHArr  []string
	RightMTHArr []string
}

func New_FRecvPacket_MTHText() FRecvPacket_MTHText {
	result := FRecvPacket_MTHText{}
	result.PacketName = "FRecvPacket_MTHText"
	return result
}

type FRecvPacket_ModeratorPasswordInvalid struct {
	PacketName string
}

func New_FRecvPacket_ModeratorPasswordInvalid() FRecvPacket_ModeratorPasswordInvalid {
	result := FRecvPacket_ModeratorPasswordInvalid{}
	result.PacketName = "FRecvPacket_ModeratorPasswordInvalid"
	return result
}

type FRecvPacket_ModeratorNotInRoom struct {
	PacketName string
}

func New_FRecvPacket_ModeratorNotInRoom() FRecvPacket_ModeratorNotInRoom {
	result := FRecvPacket_ModeratorNotInRoom{}
	result.PacketName = "FRecvPacket_ModeratorNotInRoom"
	return result
}

type FRecvPacket_StartAudio struct {
	PacketName  string
	ModeratorId string
}

func New_FRecvPacket_StartAudio() FRecvPacket_StartAudio {
	result := FRecvPacket_StartAudio{}
	result.PacketName = "FRecvPacket_StartAudio"
	return result
}

type FRecvPacket_StopAudio struct {
	PacketName  string
	ModeratorId string
}

func New_FRecvPacket_StopAudio() FRecvPacket_StopAudio {
	result := FRecvPacket_StopAudio{}
	result.PacketName = "FRecvPacket_StopAudio"
	return result
}

type FRecvPacket_NewPlayerJoinRoom struct {
	PacketName string
	PlayerId   string
}

func New_FRecvPacket_NewPlayerJoinRoom() FRecvPacket_NewPlayerJoinRoom {
	result := FRecvPacket_NewPlayerJoinRoom{}
	result.PacketName = "FRecvPacket_NewPlayerJoinRoom"
	return result
}

type FRecvPacket_SyncAudioForNewPlayer struct {
	PacketName  string
	ModeratorId string
	AudioTime   float32
}

func New_FRecvPacket_SyncAudioForNewPlayer() FRecvPacket_SyncAudioForNewPlayer {
	result := FRecvPacket_SyncAudioForNewPlayer{}
	result.PacketName = "FRecvPacket_SyncAudioForNewPlayer"
	return result
}

type FRecvPacket_SetVoiceVolume struct {
	PacketName string
	Id         string
	Volume     float32
}

func New_FRecvPacket_SetVoiceVolume() FRecvPacket_SetVoiceVolume {
	result := FRecvPacket_SetVoiceVolume{}
	result.PacketName = "FRecvPacket_SetVoiceVolume"
	return result
}

////////// ---------- Dynamo Packet ---------- //////////

type FRecvPacket_DBSignup struct {
	PacketName string
	CarNum     string
	Status     bool
}

func New_FRecvPacket_DBSignup() FRecvPacket_DBSignup {
	result := FRecvPacket_DBSignup{}
	result.PacketName = "FRecvPacket_DBSignup"
	return result
}

type FRecvPacket_DBSignin struct {
	PacketName string

	Status             bool
	IsCostumed         bool
	IsTutorial         bool
	IsTutorialRewarded bool

	FirstName string
	LastName  string
	CarNum    string

	IsMan          bool
	TopIndex       int32
	BottomIndex    int32
	HairIndex      int32
	HairColorIndex int32
	ShoesIndex     int32
	SkinIndex      int32
	FaceIndex      int32
	AccessoryIndex int32

	HCoin      int32
	TotalHCoin int32

	ItemListTop    []bool
	ItemListBottom []bool
	ItemListShoes  []bool
	ItemListAcce   []bool

	Quest_Overview    []bool
	Quest_Brand       []bool
	Quest_Product     []bool
	Quest_LiveStation []bool
	Quest_RHQ         []bool
	Quest_Fuel        []bool

	LastLoginDay   int32
	Wrc_Count      int32
	G80ev_Count    int32
	Ioniq_Count    int32
	Elevate_Count  int32
	Minigame_Count int32

	Team string
}

func New_FRecvPacket_DBSignin() FRecvPacket_DBSignin {
	result := FRecvPacket_DBSignin{}
	result.PacketName = "FRecvPacket_DBSignin"
	return result
}

type FRecvPacket_LoveForest struct {
	PacketName string
	Progress   map[string]float32
}

func New_FRecvPacket_LoveForest() FRecvPacket_LoveForest {
	result := FRecvPacket_LoveForest{}
	result.PacketName = "FRecvPacket_LoveForest"
	return result
}

type FRecvPacket_HCoinRank struct {
	PacketName string
	Rank       int32
}

func New_FRecvPacket_HCoinRank() FRecvPacket_HCoinRank {
	result := FRecvPacket_HCoinRank{}
	result.PacketName = "FRecvPacket_HCoinRank"
	return result
}

////////// ---------- Dynamo Packet ---------- //////////

type FRecvPacket_PlayerLevelChange struct {
	PacketName string
	LevelName  string
}

func New_FRecvPacket_PlayerLevelChange() FRecvPacket_PlayerLevelChange {
	result := FRecvPacket_PlayerLevelChange{}
	result.PacketName = "FRecvPacket_PlayerLevelChange"
	return result
}

type FRecvPacket_CheckAliveOtherChar struct {
	PacketName string
	Id         string
}

func New_FRecvPacket_ServerUserCount() FRecvPacket_ServerUserCount {
	result := FRecvPacket_ServerUserCount{}
	result.PacketName = "FRecvPacket_ServerUserCount"
	return result
}

type FRecvPacket_ServerUserCount struct {
	PacketName  string
	UserCount   int32
	ServerIndex int32
}

func New_FRecvPacket_CheckAliveOtherChar() FRecvPacket_CheckAliveOtherChar {
	result := FRecvPacket_CheckAliveOtherChar{}
	result.PacketName = "FRecvPacket_CheckAliveOtherChar"
	return result
}

// Dummy

type FRecvPacket_DummyPosList struct {
	PacketName string
	XPositions []int32
	YPositions []int32
}

func New_FRecvPacket_DummyPosList() FRecvPacket_DummyPosList {
	result := FRecvPacket_DummyPosList{}
	result.PacketName = "FRecvPacket_DummyPosList"
	return result
}

type FRecvPacket_DummyPosSet struct {
	PacketName string
	ListIndex  int32
	XPosition  int32
	YPosition  int32
}

func New_FRecvPacket_DummyPosSet() FRecvPacket_DummyPosSet {
	result := FRecvPacket_DummyPosSet{}
	result.PacketName = "FRecvPacket_DummyPosSet"
	return result
}

// Squid Game

type FRecvPacket_SquidGameEnter struct {
	PacketName string
	State      bool
}

func New_FRecvPacket_SquidGameEnter() FRecvPacket_SquidGameEnter {
	result := FRecvPacket_SquidGameEnter{}
	result.PacketName = "FRecvPacket_SquidGameEnter"
	return result
}

type FRecvPacket_SquidGameStart struct {
	PacketName string
	SoundSpeed []int32
	DummyFate  []int32
	DieSpeed   []int32
}

func New_FRecvPacket_SquidGameStart() FRecvPacket_SquidGameStart {
	result := FRecvPacket_SquidGameStart{}
	result.PacketName = "FRecvPacket_SquidGameStart"
	return result
}

type FRecvPacket_SquidGameDie struct {
	PacketName string
	//Signal     byte
}

func New_FRecvPacket_SquidGameDie() FRecvPacket_SquidGameDie {
	result := FRecvPacket_SquidGameDie{}
	result.PacketName = "FRecvPacket_SquidGameDie"
	return result
}

type FRecvPacket_TimeCheck struct {
	PacketName string
	Hour       int32
	Minute     int32
	Second     int32
}

func New_FRecvPacket_TimeCheck() FRecvPacket_TimeCheck {
	result := FRecvPacket_TimeCheck{}
	result.PacketName = "FRecvPacket_TimeCheck"
	return result
}

type FRecvPacket_SquidGameLogin struct {
	PacketName     string
	Id             string
	Position       transform.Vector3
	Rotation       transform.Vector3
	IsMan          bool
	SkinIndex      int32
	TopIndex       int32
	BottomIndex    int32
	HairIndex      int32
	ShoesIndex     int32
	HairColorIndex int32
	FaceIndex      int32
	AccessoryIndex int32
	FirstName      string
	LastName       string
	BirthDay       int32
	BirthMonth     int32
	DealerType     string
	Country        string
	RoomIndex      int32
}

func New_FRecvPacket_SquidGameLogin() FRecvPacket_SquidGameLogin {
	result := FRecvPacket_SquidGameLogin{}
	result.PacketName = "FRecvPacket_SquidGameLogin"
	return result
}

type FRecvPacket_SquidRefresh struct {
	PacketName    string
	RoomPlayerNum []int32
}

func New_FRecvPacket_SquidRefresh() FRecvPacket_SquidRefresh {
	result := FRecvPacket_SquidRefresh{}
	result.PacketName = "FRecvPacket_SquidRefresh"
	return result
}

type FRecvPacket_SquidPlayerNum struct {
	PacketName string
	PlayerNum  int32
}

func New_FRecvPacket_SquidPlayerNum() FRecvPacket_SquidPlayerNum {
	result := FRecvPacket_SquidPlayerNum{}
	result.PacketName = "FRecvPacket_SquidPlayerNum"
	return result
}

// Friend System

type FRecvPacket_FriendRequest struct {
	PacketName string
	TargetId   string
	Name       string
}

func New_FRecvPacket_FriendRequest() FRecvPacket_FriendRequest {
	result := FRecvPacket_FriendRequest{}
	result.PacketName = "FRecvPacket_FriendRequest"
	return result
}

type FRecvPacket_AcceptFriend struct {
	PacketName string
	Id         string
	TargetId   string
	BeFriend   bool
}

func New_FRecvPacket_AcceptFriend() FRecvPacket_AcceptFriend {
	result := FRecvPacket_AcceptFriend{}
	result.PacketName = "FRecvPacket_AcceptFriend"
	return result
}

type FRecvPacket_FriendList struct {
	PacketName string
	FriendList []FriendSearch
}

func New_FRecvPacket_FriendList() FRecvPacket_FriendList {
	result := FRecvPacket_FriendList{}
	result.PacketName = "FRecvPacket_FriendList"
	return result
}

type FriendSearch struct {
	Name     string
	FriendId string
	IsOnline bool
}

type FRecvPacket_FriendSearch struct {
	PacketName        string
	FriendSearchArray []FriendSearch
}

func New_FRecvPacket_FriendSearch() FRecvPacket_FriendSearch {
	result := FRecvPacket_FriendSearch{}
	result.PacketName = "FRecvPacket_FriendSearch"
	return result
}

type FRecvPacket_FriendInfo struct {
	PacketName string
	Name       string
	Team       string
	FriendId   string
	Message    string
	FriendNum  int32
}

func New_FRecvPacket_FriendInfo() FRecvPacket_FriendInfo {
	result := FRecvPacket_FriendInfo{}
	result.PacketName = "FRecvPacket_FriendInfo"
	return result
}
