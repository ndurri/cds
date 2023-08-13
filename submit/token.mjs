import { s3fetch, s3put } from "./s3.mjs"
import { fetch, urlEncode } from "./fetch.mjs"

const tokenURL = process.env.TOKEN_URL
const clientId = process.env.CLIENT_ID
const clientSecret = process.env.CLIENT_SECRET
const tokenBucket = process.env.TOKEN_BUCKET

export const getToken = async id => {
	const resp = await s3fetch(tokenBucket, "tokens", id)
	const t = await resp.json()
	const expires = new Date(resp.lastModified)
	expires.setSeconds(expires.getSeconds() + t.expires_in)
	const rt = {}
	rt.id = id
	rt.access_token = t.access_token
	rt.expired = () => expires < new Date()
	rt.refresh = async () => refresh(rt, t.refresh_token)
	rt.save = () => {}

	return rt
}

const refresh = async (t, refresh_token) => {
	if(!t.expired()) return t
	const params = {
		client_id:     clientId,
		client_secret: clientSecret,
		grant_type:    "refresh_token",
		refresh_token: refresh_token
	}
	const options = {
		method: "POST",
		headers: {
			"Content-Type": "application/x-www-form-urlencoded"
		},
		body: urlEncode(params)
	}
	const t2 = await fetch(tokenURL, options).then(res => res.json())
	const expires = new Date()
	expires.setSeconds(expires.getSeconds() + t2.expires_in)
	const rt = {}
	rt.id = t.id
	rt.access_token = t2.access_token
	rt.expired = () => expires < new Date()
	rt.refresh = async () => refresh(rt, t2.refresh_token)
	rt.save = () => save(t.id, t2)

	return rt
}

const save = (id, t) => {
	return s3put(tokenBucket, "tokens", id, JSON.stringify(t))
}