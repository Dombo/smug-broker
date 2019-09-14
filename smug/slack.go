package smug


import (
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    libsl "github.com/nlopes/slack"
)


type SlackUser struct {
    Id string
    Nick string
    Avatar string
}


type SlackUserCache struct {
    users map[string]*SlackUser
}


func (suc *SlackUserCache) Username(sb *SlackBroker, ukey string) string {
    if val, ok := suc.users[ukey]; ok {
        return val.Nick
    }
    user, err := sb.api.GetUserInfo(ukey)
    if err != nil {
        return "piglet"
    } else {
        suc.users[ukey] = &SlackUser{
            Id: ukey,
            Nick: user.Name,
            Avatar: user.Profile.Image72,
        }
    }
    fmt.Println(suc.users[ukey])
    return user.Name
}


type SlackBroker struct {
    // components from slack lib
    api *libsl.Client
    rtm *libsl.RTM
    // internal plumbing
    usercache *SlackUserCache
    chanid string
    channel string
    token string
}


func (sb *SlackBroker) CachedUsername(ukey string) string {
    return sb.usercache.Username(sb,ukey)
}


func (sb *SlackBroker) Name() string {
    return fmt.Sprintf("slack-%s", sb.channel)
}

// args [token, channel]
func (sb *SlackBroker) Setup(args ...string) {
    sb.usercache = &SlackUserCache{}
    sb.usercache.users = make(map[string]*SlackUser)
    sb.token = args[0]
    sb.channel = args[1]
    sc := libsl.New(
        sb.token,
        libsl.OptionDebug(true),
        libsl.OptionLog(log.New(
            os.Stdout,
            "smugbot: ",
            log.Lshortfile|log.LstdFlags ),
        ),
    )
    sb.api = sc
    sb.rtm = sb.api.NewRTM()

    channels, err := sb.api.GetChannels(false)
    if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
    for _, channel := range channels {
        if channel.Name == sb.channel {
            sb.chanid = channel.ID
            fmt.Println("Channel is ", channel.Name, " with id:", sb.chanid)
            break
        }
	}
    if sb.chanid == "" {
        fmt.Println("channel not found.  case?")
        return
    }

}


func (sb *SlackBroker) Put(msg string) {
    sb.rtm.SendMessage(sb.rtm.NewOutgoingMessage(msg, sb.chanid))
}


func (sb *SlackBroker) Publish(ev *Event) {
    sb.api.PostMessage(
        sb.chanid,
        libsl.MsgOptionText(ev.Text, false),
        libsl.MsgOptionUsername(ev.Nick),
        libsl.MsgOptionIconEmoji(fmt.Sprintf(":avatar_%s:", ev.Nick)),
    )
    // sb.rtm.SendMessage(sb.rtm.NewOutgoingMessage(msg, sb.chanid))
    // sb.Put(fmt.Sprintf("%s| %s", ev.Nick, ev.Text))
}


func (sb *SlackBroker) Run(dis Dispatcher) {
    if sb.rtm == nil {
        // raise some error here XXX TODO
    }
    go sb.rtm.ManageConnection()
    for msg := range sb.rtm.IncomingEvents {
        switch e := msg.Data.(type) {
        case *libsl.HelloEvent:
            // ignore Hello
        case *libsl.ConnectedEvent:
            fmt.Println("joining chan:", sb.channel)
            sb.Put("hi")
        case *libsl.MessageEvent:
            // smugbot: 2019/09/14 08:47:44 websocket_managed_conn.go:369:
            // Incoming Event:
            // {"client_msg_id":"ed722fbc-5b37-4f78-9981-e3c9ce5c85a1","suppress_notification":false,"type":"message","text":"test","user":"U6CRHMXK4","team":"T6CRHMX5G","user_team":"T6CRHMX5G","source_team":"T6CRHMX5G","channel":"C6MR9CBGR","event_ts":"1568468854.004200","ts":"1568468854.004200"}
            fmt.Printf("XYZXXX Message: %+v\n", e)
            if e.BotID == "" {
                outmsgs := []string{e.Text}
                if len(e.Files) > 0 {
                    for _,f := range e.Files {
                        outmsgs = append(outmsgs,
                            fmt.Sprintf("%s(%s)",f.Name,f.URLPrivate))
                    }
                }
                ev := &Event{
                    Origin: sb,
                    Nick: sb.CachedUsername(e.User),
                    Text: strings.Join(outmsgs, " "),
                    ts: time.Now(),
                }
                dis.Broadcast(ev)
            }
        case *libsl.PresenceChangeEvent:
            fmt.Printf("Presence Change: %v\n", e)
        case *libsl.LatencyReport:
            fmt.Printf("Current latency: %v\n", e.Value)
        case *libsl.RTMError:
            fmt.Printf("Error: %s\n", e.Error())
        case *libsl.InvalidAuthEvent:
            fmt.Printf("Invalid credentials")
            return
        default:
            // Ignore other events..
            // fmt.Printf("Unexpected: %v\n", msg.Data)
        }
    }
}


