package tg

// BotToMTProtoID Bot to MTProto ID.
func BotToMTProtoID(id int32) int64 {
	// Telegram Bot API looks like -1001129109101
	return -(int64(id) % -1000000000000)
}

// MTProtoToBotID converts MTProto to Bot ID
func MTProtoToBotID(id int32) int64 {
	return -(int64(id) + 1000000000000)
}
