package tg

const idOffset = 1000000000000

// BotToMTProtoID Bot to MTProto ID.
func BotToMTProtoID(id int64) int64 {
	// Telegram Bot API looks like -1001129109101
	return -(id % -idOffset)
}

// MTProtoToBotID converts MTProto to Bot ID
func MTProtoToBotID(id int32) int64 {
	return -(int64(id) + idOffset)
}
