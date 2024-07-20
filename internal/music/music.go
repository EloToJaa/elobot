package music

import (
	"fmt"
	"io"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jogramming/dca"
	log "github.com/sirupsen/logrus"
)

type BotStatus int32

const (
	Resting BotStatus = 0
	Playing BotStatus = 1
	Paused  BotStatus = 2
	Err     BotStatus = 3
)

type MuzikasBot struct {
	session       *discordgo.Session
	voiceConn     *discordgo.VoiceConnection
	queue         chan *Song
	queueList     []string
	skipInterrupt chan bool
	currentStream *dca.StreamingSession
	botStatus     BotStatus
}

func NewMuzikasBot(session *discordgo.Session) *MuzikasBot {
	return &MuzikasBot{
		session:       session,
		queue:         make(chan *Song, 100),
		skipInterrupt: make(chan bool, 1),
		botStatus:     Resting,
	}
}

func (mb *MuzikasBot) PlaySong() {
	song := mb.Dequeue()

	log.Printf("Playing song: %v \n", song.name)

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"

	encodingSession, err := dca.EncodeFile(song.downloadUrl, options)
	if err != nil {
		log.Println("Error encoding from yt url")
		log.Println(err.Error())
		return
	}
	defer encodingSession.Cleanup()

	time.Sleep(250 * time.Millisecond)
	err = mb.voiceConn.Speaking(true)

	if err != nil {
		log.Println("Error connecting to discord voice")
		log.Println(err.Error())
	}

	done := make(chan error)
	stream := dca.NewStream(encodingSession, mb.voiceConn, done)
	mb.currentStream = stream
	log.Println("Created stream, waiting on finish or err")

	mb.botStatus = Playing

	select {
	case err := <-done:
		log.Println("Song done")
		if err != nil && err != io.EOF {
			mb.botStatus = Err
			log.Println(err.Error())
			return
		}
		mb.voiceConn.Speaking(false)
		break
	case <-mb.skipInterrupt:
		log.Println("Song interrupted, stop playing")
		mb.voiceConn.Speaking(false)
		return
	}
	mb.voiceConn.Speaking(false)

	if len(mb.queue) == 0 {
		time.Sleep(250 * time.Millisecond)
		log.Println("Audio done")
		mb.Stop()
		mb.botStatus = Resting
		return
	}

	time.Sleep(250 * time.Millisecond)
	log.Println("Play next in queue")
	go mb.PlaySong()
}

func (mb *MuzikasBot) Skip() {
	if len(mb.queue) == 0 {
		mb.Stop()
	} else {
		if len(mb.skipInterrupt) == 0 {
			mb.skipInterrupt <- true
			mb.PlaySong()
		}
	}
}

func (mb *MuzikasBot) Enqueue(song *Song) {
	log.Printf("Queueing song %v", song.name)
	songString := fmt.Sprintf("-- :%v \n", song.name)
	mb.queueList = append(mb.queueList, songString)
	mb.queue <- song
}

func (mb *MuzikasBot) Dequeue() *Song {
	mb.queueList = mb.queueList[1:]
	return <-mb.queue
}

func (mb *MuzikasBot) Stop() {
	mb.voiceConn.Disconnect()
	mb.voiceConn = nil
	mb.botStatus = Resting
}

func (mb *MuzikasBot) Pause() {
	if mb.currentStream != nil {
		mb.currentStream.SetPaused(true)
		log.Println("Paused playback")
	}
}

func (mb *MuzikasBot) Unpause() {
	if mb.currentStream != nil {
		mb.currentStream.SetPaused(false)
		log.Println("Unpaused playback")
	}
}
