package core

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zekroTJA/shinpuru/pkg/multierror"

	"github.com/zekroTJA/shinpuru/internal/util"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/snowflake"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	DB *sql.DB
}

func (m *Sqlite) setup() {
	mErr := multierror.New(nil)

	_, err := m.DB.Exec("CREATE TABLE IF NOT EXISTS `guilds` (" +
		"`iid` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`guildID` text NOT NULL DEFAULT ''," +
		"`prefix` text NOT NULL DEFAULT ''," +
		"`autorole` text NOT NULL DEFAULT ''," +
		"`modlogchanID` text NOT NULL DEFAULT ''," +
		"`voicelogchanID` text NOT NULL DEFAULT ''," +
		"`muteRoleID` text NOT NULL DEFAULT ''," +
		"`ghostPingMsg` text NOT NULL DEFAULT ''," +
		"`jdoodleToken` text NOT NULL DEFAULT ''," +
		"`backup` text NOT NULL DEFAULT ''," +
		"`inviteBlock` text NOT NULL DEFAULT ''," +
		"`joinMsg` text NOT NULL DEFAULT ''," +
		"`leaveMsg` text NOT NULL DEFAULT ''" +
		");")
	mErr.Append(err)

	_, err = m.DB.Exec("CREATE TABLE IF NOT EXISTS `permissions` (" +
		"`iid` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`roleID` text NOT NULL DEFAULT ''," +
		"`guildID` text NOT NULL DEFAULT ''," +
		"`permission` int(11) NOT NULL DEFAULT '0'" +
		");")
	mErr.Append(err)

	_, err = m.DB.Exec("CREATE TABLE IF NOT EXISTS `reports` (" +
		"`iid` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`id` text NOT NULL DEFAULT ''," +
		"`type` int(11) NOT NULL DEFAULT '3'," +
		"`guildID` text NOT NULL DEFAULT ''," +
		"`executorID` text NOT NULL DEFAULT ''," +
		"`victimID` text NOT NULL DEFAULT ''," +
		"`msg` text NOT NULL DEFAULT ''," +
		"`attachment` text NOT NULL DEFAULT ''" +
		");")
	mErr.Append(err)

	_, err = m.DB.Exec("CREATE TABLE IF NOT EXISTS `settings` (" +
		"`iid` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`setting` text NOT NULL DEFAULT ''," +
		"`value` text NOT NULL DEFAULT ''" +
		");")
	mErr.Append(err)

	_, err = m.DB.Exec("CREATE TABLE IF NOT EXISTS `settings` (" +
		"`iid` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`setting` text NOT NULL DEFAULT ''," +
		"`value` text NOT NULL DEFAULT ''" +
		");")
	mErr.Append(err)

	_, err = m.DB.Exec("CREATE TABLE IF NOT EXISTS `starboard` (" +
		"`iid` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`guildID` text NOT NULL DEFAULT ''," +
		"`chanID` text NOT NULL DEFAULT ''," +
		"`enabled` tinyint(1) NOT NULL DEFAULT '1'," +
		"`minimum` int(11) NOT NULL DEFAULT '5'" +
		");")
	mErr.Append(err)

	_, err = m.DB.Exec("CREATE TABLE IF NOT EXISTS `votes` (" +
		"`iid` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`id` text NOT NULL DEFAULT ''," +
		"`data` mediumtext NOT NULL DEFAULT ''" +
		");")
	mErr.Append(err)

	_, err = m.DB.Exec("CREATE TABLE IF NOT EXISTS `twitchnotify` (" +
		"`iid` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`guildID` text NOT NULL DEFAULT ''," +
		"`channelID` text NOT NULL DEFAULT ''," +
		"`twitchUserID` text NOT NULL DEFAULT ''" +
		");")
	mErr.Append(err)

	_, err = m.DB.Exec("CREATE TABLE IF NOT EXISTS `backups` (" +
		"`iid` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`guildID` text NOT NULL DEFAULT ''," +
		"`timestamp` bigint(20) NOT NULL DEFAULT 0," +
		"`fileID` text NOT NULL DEFAULT ''" +
		");")
	mErr.Append(err)

	_, err = m.DB.Exec("CREATE TABLE IF NOT EXISTS `tags` (" +
		"`iid` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"`id` text NOT NULL DEFAULT ''," +
		"`ident` text NOT NULL DEFAULT ''," +
		"`creatorID` text NOT NULL DEFAULT ''," +
		"`guildID` text NOT NULL DEFAULT ''," +
		"`content` text NOT NULL DEFAULT ''," +
		"`created` bigint(20) NOT NULL DEFAULT 0," +
		"`lastEdit` bigint(20) NOT NULL DEFAULT 0" +
		");")
	mErr.Append(err)

	if mErr.Len() > 0 {
		util.Log.Fatalf("Failed database setup: %s", mErr.Concat().Error())
	}
}

func (m *Sqlite) Connect(credentials ...interface{}) error {
	var err error
	creds := credentials[0].(*ConfigDatabaseFile)
	if creds == nil {
		return errors.New("Database credentials from config were nil")
	}
	dsn := fmt.Sprintf("file:" + creds.DBFile)
	m.DB, err = sql.Open("sqlite3", dsn)
	m.setup()
	return err
}

func (m *Sqlite) Close() {
	if m.DB != nil {
		m.DB.Close()
	}
}

func (m *Sqlite) getGuildSetting(guildID, key string) (string, error) {
	var value string
	err := m.DB.QueryRow("SELECT "+key+" FROM guilds WHERE guildID = ?", guildID).Scan(&value)
	if err == sql.ErrNoRows {
		err = ErrDatabaseNotFound
	}
	return value, err
}

func (m *Sqlite) setGuildSetting(guildID, key string, value string) error {
	res, err := m.DB.Exec("UPDATE guilds SET "+key+" = ? WHERE guildID = ?", value, guildID)
	if err != nil {
		return err
	}
	if ar, err := res.RowsAffected(); ar == 0 {
		if err != nil {
			return err
		}
		_, err := m.DB.Exec("INSERT INTO guilds (guildID, "+key+") VALUES (?, ?)", guildID, value)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return err
}

func (m *Sqlite) GetGuildPrefix(guildID string) (string, error) {
	val, err := m.getGuildSetting(guildID, "prefix")
	return val, err
}

func (m *Sqlite) SetGuildPrefix(guildID, newPrefix string) error {
	return m.setGuildSetting(guildID, "prefix", newPrefix)
}

func (m *Sqlite) GetGuildAutoRole(guildID string) (string, error) {
	val, err := m.getGuildSetting(guildID, "autorole")
	return val, err
}

func (m *Sqlite) SetGuildAutoRole(guildID, autoRoleID string) error {
	return m.setGuildSetting(guildID, "autorole", autoRoleID)
}

func (m *Sqlite) GetGuildModLog(guildID string) (string, error) {
	val, err := m.getGuildSetting(guildID, "modlogchanID")
	return val, err
}

func (m *Sqlite) SetGuildModLog(guildID, chanID string) error {
	return m.setGuildSetting(guildID, "modlogchanID", chanID)
}

func (m *Sqlite) GetGuildVoiceLog(guildID string) (string, error) {
	val, err := m.getGuildSetting(guildID, "voicelogchanID")
	return val, err
}

func (m *Sqlite) SetGuildVoiceLog(guildID, chanID string) error {
	return m.setGuildSetting(guildID, "voicelogchanID", chanID)
}

func (m *Sqlite) GetGuildNotifyRole(guildID string) (string, error) {
	val, err := m.getGuildSetting(guildID, "notifyRoleID")
	return val, err
}

func (m *Sqlite) SetGuildNotifyRole(guildID, roleID string) error {
	return m.setGuildSetting(guildID, "notifyRoleID", roleID)
}

func (m *Sqlite) GetGuildGhostpingMsg(guildID string) (string, error) {
	val, err := m.getGuildSetting(guildID, "ghostPingMsg")
	return val, err
}

func (m *Sqlite) SetGuildGhostpingMsg(guildID, msg string) error {
	return m.setGuildSetting(guildID, "ghostPingMsg", msg)
}

func (m *Sqlite) GetMemberPermissionLevel(s *discordgo.Session, guildID string, memberID string) (int, error) {
	guildPerms, err := m.GetGuildPermissions(guildID)
	if err != nil {
		return 0, err
	}
	member, err := s.GuildMember(guildID, memberID)
	if err != nil {
		return 0, err
	}
	maxPermLvl := 0
	if lvl, ok := guildPerms[guildID]; ok {
		maxPermLvl = lvl
	}
	for _, rID := range member.Roles {
		if lvl, ok := guildPerms[rID]; ok && lvl > maxPermLvl {
			maxPermLvl = lvl
		}
	}
	return maxPermLvl, err
}

func (m *Sqlite) GetGuildPermissions(guildID string) (map[string]int, error) {
	results := make(map[string]int)
	rows, err := m.DB.Query("SELECT roleID, permission FROM permissions WHERE guildID = ?",
		guildID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var roleID string
		var permission int
		err := rows.Scan(&roleID, &permission)
		if err != nil {
			return nil, err
		}
		results[roleID] = permission
	}
	return results, nil
}

func (m *Sqlite) SetGuildRolePermission(guildID, roleID string, permLvL int) error {
	res, err := m.DB.Exec("UPDATE permissions SET permission = ? WHERE roleID = ? AND guildID = ?",
		permLvL, roleID, guildID)
	if err != nil {
		return err
	}
	if ar, err := res.RowsAffected(); ar == 0 {
		if err != nil {
			return err
		}
		_, err := m.DB.Exec("INSERT INTO permissions (roleID, guildID, permission) VALUES (?, ?, ?)",
			roleID, guildID, permLvL)
		return err
	}
	return nil
}

func (m *Sqlite) GetGuildJdoodleKey(guildID string) (string, error) {
	val, err := m.getGuildSetting(guildID, "jdoodleToken")
	return val, err
}

func (m *Sqlite) SetGuildJdoodleKey(guildID, key string) error {
	return m.setGuildSetting(guildID, "jdoodleToken", key)
}

func (m *Sqlite) GetGuildBackup(guildID string) (bool, error) {
	val, err := m.getGuildSetting(guildID, "backup")
	return val != "", err
}

func (m *Sqlite) SetGuildBackup(guildID string, enabled bool) error {
	var val string
	if enabled {
		val = "1"
	}
	return m.setGuildSetting(guildID, "backup", val)
}

func (m *Sqlite) GetSetting(setting string) (string, error) {
	var value string
	err := m.DB.QueryRow("SELECT value FROM settings WHERE setting = ?", setting).Scan(&value)
	if err == sql.ErrNoRows {
		err = ErrDatabaseNotFound
	}
	return value, err
}

func (m *Sqlite) SetSetting(setting, value string) error {
	res, err := m.DB.Exec("UPDATE settings SET value = ? WHERE setting = ?", value, setting)
	if ar, err := res.RowsAffected(); ar == 0 {
		if err != nil {
			return err
		}
		_, err := m.DB.Exec("INSERT INTO settings (setting, value) VALUES (?, ?)", setting, value)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return err
}

func (m *Sqlite) AddReport(rep *util.Report) error {
	_, err := m.DB.Exec("INSERT INTO reports (id, type, guildID, executorID, victimID, msg, attachment) VALUES (?, ?, ?, ?, ?, ?, ?)",
		rep.ID, rep.Type, rep.GuildID, rep.ExecutorID, rep.VictimID, rep.Msg, rep.AttachmehtURL)
	return err
}

func (m *Sqlite) DeleteReport(id snowflake.ID) error {
	_, err := m.DB.Exec("DELETE FROM reports WHERE id = ?", id)
	return err
}

func (m *Sqlite) GetReport(id snowflake.ID) (*util.Report, error) {
	rep := new(util.Report)

	row := m.DB.QueryRow("SELECT id, type, guildID, executorID, victimID, msg, attachment FROM reports WHERE id = ?", id)
	err := row.Scan(&rep.ID, &rep.Type, &rep.GuildID, &rep.ExecutorID, &rep.VictimID, &rep.Msg, &rep.AttachmehtURL)
	if err == sql.ErrNoRows {
		return nil, ErrDatabaseNotFound
	}

	return rep, err
}

func (m *Sqlite) GetReportsGuild(guildID string) ([]*util.Report, error) {
	rows, err := m.DB.Query("SELECT id, type, guildID, executorID, victimID, msg, attachment FROM reports WHERE guildID = ?", guildID)
	var results []*util.Report
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		rep := new(util.Report)
		err := rows.Scan(&rep.ID, &rep.Type, &rep.GuildID, &rep.ExecutorID, &rep.VictimID, &rep.Msg, &rep.AttachmehtURL)
		if err != nil {
			return nil, err
		}
		results = append(results, rep)
	}
	return results, nil
}

func (m *Sqlite) GetReportsFiltered(guildID, memberID string, repType int) ([]*util.Report, error) {
	query := fmt.Sprintf(`SELECT id, type, guildID, executorID, victimID, msg, attachment FROM reports WHERE guildID = "%s"`, guildID)
	if memberID != "" {
		query += fmt.Sprintf(` AND victimID = "%s"`, memberID)
	}
	if repType != -1 {
		query += fmt.Sprintf(` AND type = %d`, repType)
	}
	rows, err := m.DB.Query(query)
	var results []*util.Report
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		rep := new(util.Report)
		err := rows.Scan(&rep.ID, &rep.Type, &rep.GuildID, &rep.ExecutorID, &rep.VictimID, &rep.Msg, &rep.AttachmehtURL)
		if err != nil {
			return nil, err
		}
		results = append(results, rep)
	}
	return results, nil
}

func (m *Sqlite) GetVotes() (map[string]*util.Vote, error) {
	rows, err := m.DB.Query("SELECT id, data FROM votes")
	results := make(map[string]*util.Vote)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var voteID, rawData string
		err := rows.Scan(&voteID, &rawData)
		if err != nil {
			util.Log.Error("An error occured reading vote from database: ", err)
			continue
		}
		vote, err := util.VoteUnmarshal(rawData)
		if err != nil {
			m.DeleteVote(rawData)
		} else {
			results[vote.ID] = vote
		}
	}
	return results, err
}

func (m *Sqlite) AddUpdateVote(vote *util.Vote) error {
	rawData, err := vote.Marshal()
	if err != nil {
		return err
	}
	res, err := m.DB.Exec("UPDATE votes SET data = ? WHERE id = ?", rawData, vote.ID)
	if ar, err := res.RowsAffected(); ar == 0 {
		if err != nil {
			return err
		}
		_, err := m.DB.Exec("INSERT INTO votes (id, data) VALUES (?, ?)", vote.ID, rawData)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return err
}

func (m *Sqlite) DeleteVote(voteID string) error {
	_, err := m.DB.Exec("DELETE FROM votes WHERE id = ?", voteID)
	return err
}

func (m *Sqlite) GetMuteRoles() (map[string]string, error) {
	rows, err := m.DB.Query("SELECT guildID, muteRoleID FROM guilds")
	results := make(map[string]string)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var guildID, roleID string
		err = rows.Scan(&guildID, &roleID)
		if err == nil {
			results[guildID] = roleID
		}
	}
	return results, nil
}

func (m *Sqlite) GetMuteRoleGuild(guildID string) (string, error) {
	val, err := m.getGuildSetting(guildID, "muteRoleID")
	return val, err
}

func (m *Sqlite) SetMuteRole(guildID, roleID string) error {
	return m.setGuildSetting(guildID, "muteRoleID", roleID)
}

func (m *Sqlite) GetTwitchNotify(twitchUserID, guildID string) (*TwitchNotifyDBEntry, error) {
	t := &TwitchNotifyDBEntry{
		TwitchUserID: twitchUserID,
		GuildID:      guildID,
	}
	err := m.DB.QueryRow("SELECT channelID FROM twitchnotify WHERE twitchUserID = ? AND guildID = ?",
		twitchUserID, guildID).Scan(&t.ChannelID)
	if err == sql.ErrNoRows {
		err = ErrDatabaseNotFound
	}
	return t, err
}

func (m *Sqlite) SetTwitchNotify(twitchNotify *TwitchNotifyDBEntry) error {
	res, err := m.DB.Exec("UPDATE twitchnotify SET channelID = ? WHERE twitchUserID = ? AND guildID = ?",
		twitchNotify.ChannelID, twitchNotify.TwitchUserID, twitchNotify.GuildID)
	if err != nil {
		return err
	}
	if ar, err := res.RowsAffected(); ar == 0 {
		if err != nil {
			return err
		}
		_, err := m.DB.Exec("INSERT INTO twitchnotify (twitchUserID, guildID, channelID) VALUES (?, ?, ?)",
			twitchNotify.TwitchUserID, twitchNotify.GuildID, twitchNotify.ChannelID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return err
}

func (m *Sqlite) DeleteTwitchNotify(twitchUserID, guildID string) error {
	_, err := m.DB.Exec("DELETE FROM twitchnotify WHERE twitchUserID = ? AND guildID = ?", twitchUserID, guildID)
	return err
}

func (m *Sqlite) GetAllTwitchNotifies(twitchUserID string) ([]*TwitchNotifyDBEntry, error) {
	query := "SELECT twitchUserID, guildID, channelID FROM twitchnotify"
	if twitchUserID != "" {
		query += " WHERE twitchUserID = " + twitchUserID
	}
	rows, err := m.DB.Query(query)
	results := make([]*TwitchNotifyDBEntry, 0)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		t := new(TwitchNotifyDBEntry)
		err = rows.Scan(&t.TwitchUserID, &t.GuildID, &t.ChannelID)
		if err == nil {
			results = append(results, t)
		}
	}
	return results, nil
}

func (m *Sqlite) AddBackup(guildID, fileID string) error {
	timestamp := time.Now().Unix()
	_, err := m.DB.Exec("INSERT INTO backups (guildID, timestamp, fileID) VALUES (?, ?, ?)", guildID, timestamp, fileID)
	return err
}

func (m *Sqlite) DeleteBackup(guildID, fileID string) error {
	_, err := m.DB.Exec("DELETE FROM backups WHERE guildID = ? AND fileID = ?", guildID, fileID)
	return err
}

func (m *Sqlite) GetGuildInviteBlock(guildID string) (string, error) {
	return m.getGuildSetting(guildID, "inviteBlock")
}

func (m *Sqlite) SetGuildInviteBlock(guildID string, data string) error {
	return m.setGuildSetting(guildID, "inviteBlock", data)
}

func (m *Sqlite) GetGuildJoinMsg(guildID string) (string, string, error) {
	data, err := m.getGuildSetting(guildID, "joinMsg")
	if err != nil {
		return "", "", err
	}
	if data == "" {
		return "", "", nil
	}

	i := strings.Index(data, "|")
	return data[:i], data[i+1:], nil
}

func (m *Sqlite) SetGuildJoinMsg(guildID string, channelID string, msg string) error {
	return m.setGuildSetting(guildID, "joinMsg", fmt.Sprintf("%s|%s", channelID, msg))
}

func (m *Sqlite) GetGuildLeaveMsg(guildID string) (string, string, error) {
	data, err := m.getGuildSetting(guildID, "leaveMsg")
	if err != nil {
		return "", "", err
	}
	if data == "" {
		return "", "", nil
	}

	i := strings.Index(data, "|")
	return data[:i], data[i+1:], nil
}

func (m *Sqlite) SetGuildLeaveMsg(guildID string, channelID string, msg string) error {
	return m.setGuildSetting(guildID, "leaveMsg", fmt.Sprintf("%s|%s", channelID, msg))
}

func (m *Sqlite) GetBackups(guildID string) ([]*BackupEntry, error) {
	rows, err := m.DB.Query("SELECT guildID, timestamp, fileID FROM backups WHERE guildID = ?", guildID)
	if err == sql.ErrNoRows {
		return nil, ErrDatabaseNotFound
	}
	if err != nil {
		return nil, err
	}

	backups := make([]*BackupEntry, 0)
	for rows.Next() {
		be := new(BackupEntry)
		var timeStampUnix int64
		err = rows.Scan(&be.GuildID, &timeStampUnix, &be.FileID)
		if err != nil {
			return nil, err
		}
		be.Timestamp = time.Unix(timeStampUnix, 0)
		backups = append(backups, be)
	}

	return backups, nil
}

func (m *Sqlite) GetBackupGuilds() ([]string, error) {
	rows, err := m.DB.Query("SELECT guildID FROM guilds WHERE backup = '1'")
	if err == sql.ErrNoRows {
		return nil, ErrDatabaseNotFound
	}
	if err != nil {
		return nil, err
	}

	guilds := make([]string, 0)
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			return nil, err
		}
		guilds = append(guilds, s)
	}

	return guilds, err
}

func (m *Sqlite) AddTag(tag *util.Tag) error {
	_, err := m.DB.Exec("INSERT INTO tags (id, ident, creatorID, guildID, content, created, lastEdit) VALUES "+
		"(?, ?, ?, ?, ?, ?, ?)", tag.ID, tag.Ident, tag.CreatorID, tag.GuildID, tag.Content, tag.Created.Unix(), tag.LastEdit.Unix())
	return err
}

func (m *Sqlite) EditTag(tag *util.Tag) error {
	_, err := m.DB.Exec("UPDATE tags SET "+
		"ident = ?, creatorID = ?, guildID = ?, content = ?, created = ?, lastEdit = ? "+
		"WHERE id = ?", tag.Ident, tag.CreatorID, tag.GuildID, tag.Content, tag.Created.Unix(), tag.LastEdit.Unix(), tag.ID)
	if err == sql.ErrNoRows {
		return ErrDatabaseNotFound
	}
	return err
}

func (m *Sqlite) GetTagByID(id snowflake.ID) (*util.Tag, error) {
	tag := new(util.Tag)
	var timestampCreated int64
	var timestampLastEdit int64

	row := m.DB.QueryRow("SELECT id, ident, creatorID, guildID, content, created, lastEdit FROM tags "+
		"WHERE id = ?", id)

	err := row.Scan(&tag.ID, &tag.Ident, &tag.CreatorID, &tag.GuildID,
		&tag.Content, &timestampCreated, &timestampLastEdit)
	if err == sql.ErrNoRows {
		return nil, ErrDatabaseNotFound
	}
	if err != nil {
		return nil, err
	}

	tag.Created = time.Unix(timestampCreated, 0)
	tag.LastEdit = time.Unix(timestampLastEdit, 0)

	return tag, nil
}

func (m *Sqlite) GetTagByIdent(ident string, guildID string) (*util.Tag, error) {
	tag := new(util.Tag)
	var timestampCreated int64
	var timestampLastEdit int64

	row := m.DB.QueryRow("SELECT id, ident, creatorID, guildID, content, created, lastEdit FROM tags "+
		"WHERE ident = ? AND guildID = ?", ident, guildID)

	err := row.Scan(&tag.ID, &tag.Ident, &tag.CreatorID, &tag.GuildID,
		&tag.Content, &timestampCreated, &timestampLastEdit)
	if err == sql.ErrNoRows {
		return nil, ErrDatabaseNotFound
	}
	if err != nil {
		return nil, err
	}

	tag.Created = time.Unix(timestampCreated, 0)
	tag.LastEdit = time.Unix(timestampLastEdit, 0)

	return tag, nil
}

func (m *Sqlite) GetGuildTags(guildID string) ([]*util.Tag, error) {
	rows, err := m.DB.Query("SELECT id, ident, creatorID, guildID, content, created, lastEdit FROM tags "+
		"WHERE guildID = ?", guildID)
	if err == sql.ErrNoRows {
		return nil, ErrDatabaseNotFound
	}
	if err != nil {
		return nil, err
	}

	tags := make([]*util.Tag, 0)
	var timestampCreated int64
	var timestampLastEdit int64
	for rows.Next() {
		tag := new(util.Tag)
		err = rows.Scan(&tag.ID, &tag.Ident, &tag.CreatorID, &tag.GuildID,
			&tag.Content, &timestampCreated, &timestampLastEdit)
		if err != nil {
			return nil, err
		}
		tag.Created = time.Unix(timestampCreated, 0)
		tag.LastEdit = time.Unix(timestampLastEdit, 0)
		tags = append(tags, tag)
	}

	return tags, nil
}

func (m *Sqlite) DeleteTag(id snowflake.ID) error {
	_, err := m.DB.Exec("DELETE FROM tags WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return ErrDatabaseNotFound
	}
	return err
}
