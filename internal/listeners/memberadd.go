package listeners

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shinpuru/internal/core"
	"github.com/zekroTJA/shinpuru/internal/util"
)

type ListenerMemberAdd struct {
	db core.Database
}

func NewListenerMemberAdd(db core.Database) *ListenerMemberAdd {
	return &ListenerMemberAdd{
		db: db,
	}
}

func (l *ListenerMemberAdd) Handler(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	autoRoleID, err := l.db.GetGuildAutoRole(e.GuildID)
	if err != nil && !core.IsErrDatabaseNotFound(err) {
		util.Log.Errorf("Failed getting autorole for guild '%s' from database: %s", e.GuildID, err.Error())
	}
	if autoRoleID != "" {
		err = s.GuildMemberRoleAdd(e.GuildID, e.User.ID, autoRoleID)
		if err != nil && strings.Contains(err.Error(), `{"code": 10011, "message": "Unknown Role"}`) {
			l.db.SetGuildAutoRole(e.GuildID, "")
		} else if err != nil {
			util.Log.Errorf("Failed setting autorole for member '%s': %s", e.User.ID, err.Error())
		}
	}

	chanID, msg, err := l.db.GetGuildJoinMsg(e.GuildID)
	if err == nil && msg != "" && chanID != "" {
		msg = strings.Replace(msg, "[user]", e.User.Username, -1)
		msg = strings.Replace(msg, "[ment]", e.User.Mention(), -1)
		util.SendEmbed(s, chanID, msg, "", 0)
	}
}
