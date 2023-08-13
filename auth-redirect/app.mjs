import * as https from 'https'
import { S3Client, PutObjectCommand, GetObjectCommand, DeleteObjectCommand } from "@aws-sdk/client-s3"

const bucket = process.env.APPDATA_BUCKET
const userBucket = process.env.USERDATA_BUCKET
const clientId = process.env.CLIENT_ID
const clientSecret = process.env.CLIENT_SECRET
const redirectURI = process.env.REDIRECT_URI
const tokenURL = process.env.TOKEN_URL

const client = new S3Client()

const body = stream => new Promise((resolve, reject) => {
	let buf = Buffer.alloc(0)
	stream.on('data', chunk => buf = Buffer.concat([buf, chunk]))
	stream.once('end', () => resolve(buf.toString('utf8')))
	stream.once('error', err => reject(err))
})

const session = async sid => {
	const params = {
		Bucket: bucket,
		Key: `authSession/${sid}`
	}
	const res = await client.send(new GetObjectCommand(params))
	const content = await body(res.Body)
	return {
		content: content,
		delete: () => client.send(new DeleteObjectCommand(params))
	}
}

const save = (key, body) => {
	const params = {
		Bucket: userBucket,
		Key: key,
		Body: body
	}
	return client.send(new PutObjectCommand(params))
}

const urlencode = obj => {
	const params = []
	for(const [key, value] of Object.entries(obj)) {
		params.push(`${encodeURI(key)}=${encodeURI(value)}`)
	}
	return params.join("&")
}

const requesttoken = code => new Promise((resolve, reject) => {
	const params = {
		"client_id":     clientId,
		"client_secret": clientSecret,
		"grant_type":    "authorization_code",
		"redirect_uri":  redirectURI,
		"code":          code
	}
	const options = {method: "POST", headers: {"content-type": "application/x-www-form-urlencoded"}}
	const req = https.request(tokenURL, options, res => {
		if(res.statusCode === 200)
			resolve(res)
		else
			reject(res)
	})
	req.once('error', err => reject(err))
	req.end(urlencode(params))
})

export const handler = async event => {
	const { state, code } = event.queryStringParameters
	const sesh = await session(state)
	const res = await requesttoken(code)
	const token = await body(res)
	await save(`tokens/${sesh.content}`, token)
	await sesh.delete()
	return {
		statusCode: 200,
		body: "Thankyou for authorizing."
	}
}