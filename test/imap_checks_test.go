// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"strconv"
	"time"

	"github.com/cucumber/godog"
	"github.com/emersion/go-imap"
)

func IMAPChecksFeatureContext(s *godog.Suite) {
	s.Step(`^IMAP response is "([^"]*)"$`, imapResponseIs)
	s.Step(`^IMAP response to "([^"]*)" is "([^"]*)"$`, imapResponseNamedIs)
	s.Step(`^IMAP response contains "([^"]*)"$`, imapResponseContains)
	s.Step(`^IMAP response to "([^"]*)" contains "([^"]*)"$`, imapResponseNamedContains)
	s.Step(`^IMAP response has (\d+) message(?:s)?$`, imapResponseHasNumberOfMessages)
	s.Step(`^IMAP response to "([^"]*)" has (\d+) message(?:s)?$`, imapResponseNamedHasNumberOfMessages)
	s.Step(`^IMAP client receives update marking message "([^"]*)" as read within (\d+) seconds$`, imapClientReceivesUpdateMarkingMessagesAsReadWithin)
	s.Step(`^IMAP client "([^"]*)" receives update marking message "([^"]*)" as read within (\d+) seconds$`, imapClientNamedReceivesUpdateMarkingMessagesAsReadWithin)
	s.Step(`^IMAP client receives update marking message "([^"]*)" as unread within (\d+) seconds$`, imapClientReceivesUpdateMarkingMessagesAsUnreadWithin)
	s.Step(`^IMAP client "([^"]*)" receives update marking message "([^"]*)" as unread within (\d+) seconds$`, imapClientNamedReceivesUpdateMarkingMessagesAsUnreadWithin)
	s.Step(`^IMAP client "([^"]*)" does not receive update for message "([^"]*)" within (\d+) seconds$`, imapClientDoesNotReceiveUpdateForMessageWithin)
}

func imapResponseIs(expectedResponse string) error {
	return imapResponseNamedIs("imap", expectedResponse)
}

func imapResponseNamedIs(clientID, expectedResponse string) error {
	res := ctx.GetIMAPLastResponse(clientID)
	if expectedResponse == "OK" {
		res.AssertOK()
	} else {
		res.AssertError(expectedResponse)
	}
	return ctx.GetTestingError()
}

func imapResponseContains(expectedResponse string) error {
	return imapResponseNamedContains("imap", expectedResponse)
}

func imapResponseNamedContains(clientID, expectedResponse string) error {
	res := ctx.GetIMAPLastResponse(clientID)
	res.AssertSections(expectedResponse)
	return ctx.GetTestingError()
}

func imapResponseHasNumberOfMessages(expectedCount int) error {
	return imapResponseNamedHasNumberOfMessages("imap", expectedCount)
}

func imapResponseNamedHasNumberOfMessages(clientID string, expectedCount int) error {
	res := ctx.GetIMAPLastResponse(clientID)
	res.AssertSectionsCount(expectedCount)
	return ctx.GetTestingError()
}

func imapClientReceivesUpdateMarkingMessagesAsReadWithin(messageUIDs string, seconds int) error {
	return imapClientNamedReceivesUpdateMarkingMessagesAsReadWithin("imap", messageUIDs, seconds)
}

func imapClientNamedReceivesUpdateMarkingMessagesAsReadWithin(clientID, messageUIDs string, seconds int) error {
	regexps := []string{}
	iterateOverSeqSet(messageUIDs, func(messageUID string) {
		regexps = append(regexps, `FETCH \(FLAGS \(.*\\Seen.*\) UID `+messageUID)
	})
	ctx.GetIMAPLastResponse(clientID).WaitForSections(time.Duration(seconds)*time.Second, regexps...)
	return ctx.GetTestingError()
}

func imapClientReceivesUpdateMarkingMessagesAsUnreadWithin(messageUIDs string, seconds int) error {
	return imapClientNamedReceivesUpdateMarkingMessagesAsUnreadWithin("imap", messageUIDs, seconds)
}

func imapClientNamedReceivesUpdateMarkingMessagesAsUnreadWithin(clientID, messageUIDs string, seconds int) error {
	regexps := []string{}
	iterateOverSeqSet(messageUIDs, func(messageUID string) {
		// Golang does not support negative look ahead. Following complex regexp checks \Seen is not there.
		regexps = append(regexps, `FETCH \(FLAGS \(([^S]|S[^e]|Se[^e]|See[^n])*\) UID `+messageUID)
	})
	ctx.GetIMAPLastResponse(clientID).WaitForSections(time.Duration(seconds)*time.Second, regexps...)
	return ctx.GetTestingError()
}

func imapClientDoesNotReceiveUpdateForMessageWithin(clientID, messageUIDs string, seconds int) error {
	regexps := []string{}
	iterateOverSeqSet(messageUIDs, func(messageUID string) {
		regexps = append(regexps, `FETCH.*UID `+messageUID)
	})
	ctx.GetIMAPLastResponse(clientID).WaitForNotSections(time.Duration(seconds)*time.Second, regexps...)
	return ctx.GetTestingError()
}

func iterateOverSeqSet(seqSet string, callback func(string)) {
	seq, err := imap.NewSeqSet(seqSet)
	if err != nil {
		panic(err)
	}
	for _, set := range seq.Set {
		for i := set.Start; i <= set.Stop; i++ {
			callback(strconv.Itoa(int(i)))
		}
	}
}
