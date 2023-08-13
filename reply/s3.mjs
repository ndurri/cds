import { S3Client, GetObjectCommand } from "@aws-sdk/client-s3"

const client = new S3Client()

const body = stream => new Promise((resolve, reject) => {
	let buf = Buffer.alloc(0)
	stream.on('data', chunk => buf = Buffer.concat([buf, chunk]))
	stream.once('end', () => resolve(buf))
	stream.once('error', err => reject(err))
})

const response = res => {
	const resp = {}
	resp.lastModified = new Date(res.LastModified)
	resp.body = res.Body
	resp.arrayBuffer = () => body(resp.body)
	resp.text = () => resp.arrayBuffer().then(buf => buf.toString('utf8'))
	resp.json = () => resp.text().then(text => JSON.parse(text))
	return resp
}

export const s3fetch = (bucket, folder, uuid) => {
	const params = {
		Bucket: bucket,
		Key: `${folder}/${uuid}`
	}
	return client.send(new GetObjectCommand(params))
		.then(response)
}