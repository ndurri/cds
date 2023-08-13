import { request } from 'node:https'

const body = res => new Promise((resolve, reject) => {
	let buf = Buffer.alloc(0)
	res.on('data', chunk => buf = Buffer.concat([buf, chunk]))
	res.once('end', () => resolve(buf))
	res.once('error', err => reject(err))
})

const response = res => {
	const resp = {}
	resp.status = res.statusCode
	resp.headers = res.headers
	resp.ok = resp.status >= 200 && resp.status <= 299
	resp.arrayBuffer = () => body(res)
	resp.text = () => resp.arrayBuffer().then(buf => buf.toString('utf8'))
	resp.json = () => resp.text().then(text => JSON.parse(text))
	return resp
}

export const fetch = (url, options) => new Promise((resolve, reject) => {
	const req = request(url, options, res => {
		resolve(response(res))
	})
	for(const [key, value] of Object.entries(options.headers)) {
		req.setHeader(key, value)
	}	
	req.once('error', err => reject(err))
	req.end(options.body)
})

export const urlEncode = obj => {
	const params = []
	for(const [key, value] of Object.entries(obj)) {
		params.push(`${encodeURI(key)}=${encodeURI(value)}`)
	}
	return params.join("&")
}