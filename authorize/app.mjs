import { URL } from "node:url"
import { S3Client, PutObjectCommand } from "@aws-sdk/client-s3"
const { randomUUID } = await import("node:crypto")

const bucket = process.env.APPDATA_BUCKET
const authURL = process.env.AUTH_URL
const clientId = process.env.CLIENT_ID
const scope = process.env.SCOPE
const redirectURI = process.env.REDIRECT_URI

const client = new S3Client()

const session = async body => {
	const sid = randomUUID({disableEntropyCache: true})
	const params = {
		Bucket: bucket,
		Key: `authSession/${sid}`,
		Body: body
	}
	await client.send(new PutObjectCommand(params))
	return sid
}

export const handler = async event => {
	let submitter = "default"
	if (event.queryStringParameters && event.queryStringParameters.submitter)
		submitter = event.queryStringParameters.submitter
	const sid = await session(submitter)
	const location = new URL(authURL)
	location.searchParams.set("response_type", "code")
	location.searchParams.set("client_id", clientId)
	location.searchParams.set("scope", scope)
	location.searchParams.set("redirect_uri", redirectURI)
	location.searchParams.set("state", sid)
	return {
		statusCode: 302,
		headers: {
			location: location.toString()
		}
	}
}