import { S3Client, PutObjectCommand } from "@aws-sdk/client-s3"

const client = new S3Client()

const bucket = process.env.BUCKET

export const handler = async event => {
	const resUUID = event.headers["X-Conversation-ID"]
	const docUUID = event.headers["X-Correlation-ID"]
	console.log("Processing %s/%s", resUUID, docUUID)
	const body = event.isBase64Encoded ? base64Decode(event.body) : event.body
	const params = {
		Bucket: bucket,
		Key: `payload-in/${resUUID}/${docUUID}`,
		Body: body
	}
	await client.send(new PutObjectCommand(params))
	return {statusCode: 200}
}

const base64Decode = encoded => Buffer.from(encoded, 'base64').toString('utf8')