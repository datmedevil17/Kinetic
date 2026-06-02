import AgoraRTC, { IAgoraRTCClient, ICameraVideoTrack, IMicrophoneAudioTrack, IAgoraRTCRemoteUser, IDataChannelConfig } from 'agora-rtc-sdk-ng'
import signal from '../signal'
import { createHash } from 'crypto'
import { generateToken } from './generateToken'

export class VideoChat {
    private client!: IAgoraRTCClient
    private microphoneTrack: IMicrophoneAudioTrack | null = null
    private cameraTrack: ICameraVideoTrack | null = null
    private currentChannel: string = ''

    private remoteUsers: { [uid: string]: IAgoraRTCRemoteUser } = {}

    private channelTimeout: NodeJS.Timeout | null = null
    private isJoining: boolean = false

    constructor() {
        if (typeof window !== 'undefined') {
            this.client = AgoraRTC.createClient({ codec: "vp8", mode: "rtc" })
            AgoraRTC.setLogLevel(4)
            this.client.on('user-published', this.onUserPublished)
            this.client.on('user-unpublished', this.onUserUnpublished)
            this.client.on('user-left', this.onUserLeft)
            this.client.on('user-info-updated', this.onUserInfoUpdated)
            this.client.on('user-joined', this.onUserJoined)
        }
    }

    private onUserInfoUpdated = (uid: string) => {
        if (!this.remoteUsers[uid]) return
        signal.emit('user-info-updated', this.remoteUsers[uid])
    }

    private onUserJoined = (user: IAgoraRTCRemoteUser) => {
        this.remoteUsers[user.uid] = user
        signal.emit('user-info-updated', user)
    }

    public onUserPublished = async (user: IAgoraRTCRemoteUser, mediaType: "audio" | "video" | "datachannel", config?: IDataChannelConfig) => {
        this.remoteUsers[user.uid] = user
        await this.client.subscribe(user, mediaType)

        if (mediaType === 'audio') {
            user.audioTrack?.play()
        }

        if (mediaType === 'audio' || mediaType === 'video') {
            signal.emit('user-info-updated', user)
        }
    }

    public onUserUnpublished = (user: IAgoraRTCRemoteUser, mediaType: "audio" | "video" | "datachannel") => {
        if (mediaType === 'audio') {
            user.audioTrack?.stop()
        }
        if (mediaType === 'audio' || mediaType === 'video') {
            signal.emit('user-info-updated', user)
        }
    }

    public onUserLeft = (user: IAgoraRTCRemoteUser, reason: string) => {
        delete this.remoteUsers[user.uid]
        signal.emit('user-left', user)
    }

    public async toggleCamera() {
        if (!this.cameraTrack) {
            this.cameraTrack = await AgoraRTC.createCameraVideoTrack()
            this.cameraTrack.play('local-video')

            if (this.client.connectionState === 'CONNECTED') {
                await this.client.publish([this.cameraTrack])
            }

            return false
        }
        await this.cameraTrack.setEnabled(!this.cameraTrack.enabled)

        if (this.client.connectionState === 'CONNECTED' && this.cameraTrack.enabled) {
            await this.client.publish([this.cameraTrack])
        }

        return !this.cameraTrack.enabled
    }

    public async toggleMicrophone() {
        if (!this.microphoneTrack) {
            this.microphoneTrack = await AgoraRTC.createMicrophoneAudioTrack()

            if (this.client.connectionState === 'CONNECTED') {
                await this.client.publish([this.microphoneTrack])
            }

            return false
        }
        await this.microphoneTrack.setEnabled(!this.microphoneTrack.enabled)

        if (this.client.connectionState === 'CONNECTED' && this.microphoneTrack.enabled) {
            await this.client.publish([this.microphoneTrack])
        }

        return !this.microphoneTrack.enabled
    }

    public playVideoTrackAtElementId(elementId: string) {
        if (this.cameraTrack) {
            this.cameraTrack.play(elementId)
        }
    }

    private resetRemoteUsers() {
        this.remoteUsers = {}
        signal.emit('reset-users')
    }

    public async joinChannel(channel: string, uid: string, realmId: string) {
        if (this.channelTimeout) {
            clearTimeout(this.channelTimeout)
        }

        this.channelTimeout = setTimeout(async () => {
            if (channel === this.currentChannel || this.isJoining) return

            const uniqueChannelId = this.createUniqueChannelId(realmId, channel)
            const token = await generateToken(uniqueChannelId)
            if (!token) return

            this.isJoining = true
            try {
                if (this.client.connectionState === 'CONNECTED') {
                    await this.client.leave()
                }
                this.resetRemoteUsers()
                this.currentChannel = channel  // set before join so re-entrant calls are blocked

                await this.client.join(process.env.NEXT_PUBLIC_AGORA_APP_ID!, uniqueChannelId, token, uid)

                if (this.microphoneTrack && this.microphoneTrack.enabled) {
                    await this.client.publish([this.microphoneTrack])
                }
                if (this.cameraTrack && this.cameraTrack.enabled) {
                    await this.client.publish([this.cameraTrack])
                }
            } catch (e) {
                this.currentChannel = ''  // reset so a retry is possible
                throw e
            } finally {
                this.isJoining = false
            }
        }, 1000)
    }

    public async leaveChannel() {
        if (this.channelTimeout) {
            clearTimeout(this.channelTimeout)
        }

        this.channelTimeout = setTimeout(async () => {
            if (this.currentChannel === '') return

            if (this.client.connectionState === 'CONNECTED') {
                await this.client.leave()
                this.currentChannel = ''
            }
            this.resetRemoteUsers()
        }, 1000)
        
    }

    public destroy() {
        if (this.cameraTrack) {
            this.cameraTrack.stop()
            this.cameraTrack.close()
        }
        if (this.microphoneTrack) {
            this.microphoneTrack.stop()
            this.microphoneTrack.close()
        }
        this.microphoneTrack = null
        this.cameraTrack = null
    }

    private createUniqueChannelId(realmId: string, channel: string): string {
        const combined = `${realmId}-${channel}`;
        return createHash('md5').update(combined).digest('hex').substring(0, 16);
    }
}

export const videoChat = new VideoChat()