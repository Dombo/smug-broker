package smug

import (
    "testing"
)

func TestConvertSlackRefs(t *testing.T) {
    sb := &SlackBroker{}
    sb.SetupInternals()

    u1 := &SlackUser{
        Id:"U6CRHMXK4",
        Nick:"aaaa",
        Avatar:"",
    }
    sb.usercache.users[u1.Id] = u1
    sb.usercache.nicks[u1.Nick] = u1

    u2 := &SlackUser{
        Id: "U54321",
        Nick: "boy",
        Avatar:"",
    }
    sb.usercache.users[u2.Id] = u2
    sb.usercache.nicks[u2.Nick] = u2


    testwants := map[string]string {
        sb.ConvertRefsToUsers(" <@U6CRHMXK4> congradulations!!!", true):
            " aaaa congradulations!!!",
        sb.ConvertRefsToUsers("<@U6CRHMXK4> dude <@U54321>", true):
            "aaaa dude boy",
        sb.ConvertRefsToUsers("<@U54321> dude <@U54321>", true):
            "boy dude boy",
        sb.ConvertRefsToUsers(" <@88888> congradulations!!!", true):
            " <@88888> congradulations!!!",
        sb.ConvertUsersToRefs("boy: gobble", true):
            "<@U54321>: gobble",
        sb.ConvertUsersToRefs("nope: hi", true):
            "nope: hi",
    }

    for want,have := range testwants {
        if want != have {
            t.Errorf("err: wanted [%s] got [%s]", want, have)
        }
    }


}

