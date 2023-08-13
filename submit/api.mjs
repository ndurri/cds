import { fetch } from "./fetch.mjs"
import { readFileSync } from "node:fs"

const apis = JSON.parse(readFileSync("./apis.json"))

export const getAPI = doctype => {
	const api = apis[doctype]
	return {
		call: (token, content) => call(api, token, content)
	}
}

const call = (api, token, content) => {
	const headers = {}
	for(const [name, value] of Object.entries(api.headers))
		headers[name] = value
	headers.authorization = `Bearer ${token}`
	const options = {
		method: "POST",
		headers: headers,
		body: content
	}
	console.log("Calling %s with access token %s.", api.endpoint, token)
	return fetch(api.endpoint, options)
}