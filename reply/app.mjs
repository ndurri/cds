import { s3fetch } from "./s3.mjs"
import mailer from 'nodemailer'

const transport = mailer.createTransport({
  host: process.env.SMTP_HOST,
  port: process.env.SMTP_PORT,
  secure: false,
  auth: {
    user: process.env.SMTP_USER,
    pass: process.env.SMTP_PASSWORD
  }
})

export const handler = async event => {
	const bucket = event.Records[0].s3.bucket.name
	const key = event.Records[0].s3.object.key
	const [ folder, respUUID, docUUID ] = key.split("/")

	const response = await s3fetch(bucket, "responses", respUUID).then(resp => resp.json())
	const content = await s3fetch(bucket, folder, `${respUUID}/${docUUID}`).then(resp => resp.text())

	const args = {
		from: process.env.SENDER,
		to: response.From,
		subject: response.Subject,
		text: content,
		html: escapeHTML(content)
	}
	await transport.sendMail(args)
	return {
		statusCode: 200
	}
}

const escapeHTML = str => str.replace(/[&<>'"]/g, 
  tag => ({
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      "'": '&#39;',
      '"': '&quot;'
    }[tag]))