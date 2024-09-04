"use strict"

class ProxyWiki {
	get(id) {
		return m.request({
			method: "GET",
			url: `//localhost:4989/api/wiki/page/${id}`,
		})
	}

	post(id, raw) {
		return m.request({
			method: "POST",
			url: `//localhost:4989/api/wiki/page/${id}`,
			body: { "raw": raw },
		})
	}

	list() {
		return m.request({
			method: "GET",
			url: `//localhost:4989/api/wiki/page/`,
		})
	}
}


class PropertyWiki {
	constructor(proxy, id) {
		this.proxy = proxy
		this.id = id
		this.raw = ""
		this.html = ""
		this.list = []
		this.editable = false
	}

	getPage() {
		this.proxy.get(this.id).then((result) => {
			this.raw = result.raw
			this.html = result.html
		})
		.catch((e) => {
			if (e.code == 404) {
				this.raw = "this is new page"
				this.html = "<p>this is new page</p>"
			} else {
				this.raw = e.response.message
			}
		})
	}

	postPage() {
		this.proxy.post(this.id, this.raw).then((result) => {
			this.getPage()
		})
		.catch((e) => {
			this.raw = e.response.message
		})
	}

	getList() {
		this.proxy.list(this.id).then((result) => {
			this.list = result.idlist
		})
		.catch((e) => {
			this.raw = e.response.message
		})
	}
}

class ViewWikipage {
	constructor(prop) {
		this.prop = prop
	}

	oninit() {
		this.prop.id = m.route.param("id")
		this.prop.getPage()
		this.prop.getList()
	}

	view() {
		if (this.prop.id == "list") {
			return [
				m("ul", [
					...this.prop.list.map((item)=>{
						return m("li", m("a", {
							onclick: () => {
								m.route.set(`/page/${item}`)
							}
						}, item))
					})
				]),
			]
		} else {
			if (this.prop.editable) {
				return [
					m("article", [
						m("header", [
							m("span", `${this.prop.id}`),
							m("div", [
								m("a", {
									onclick: () => {
										this.prop.editable = false
										this.prop.getPage()
									}
								}, "[back]"),
								m("span", " "),
								m("a", {
									onclick: () => {
										this.prop.postPage()
										this.prop.editable = false
									}
								}, "[save]"),
							]),
						]),
						m("textarea[rows=12]", {
							oninput: (e) => {
								this.prop.raw = e.target.value
							},
						}, `${this.prop.raw}`),
					]),
				]
			} else {
				return [
					m("article", [
						m("header", [
							m("span", `${this.prop.id}`),
							m("a", {
								onclick: () => {
									this.prop.editable = true
								}
							}, "[edit]"),
						]),
						m.trust(this.prop.html),
					]),
				]
			}
		}
	}
}

class ViewNavigate {
	constructor() {}

	view() {
		return [
			m("ul", [
				m("li", 
					m("h3", "Wiki"),
				),
			]),
			m("ul", [
				m("li", [
					m("a", {
						onclick: () => {
							m.route.set("/page/list")
						},
					}, "list"),
				]),
			]),
		]
	}
}


class ViewLayout {
	constructor(navi, main) {
		this.navi = navi
		this.main = main
	}

	view() {
		return [
			m("nav.container", [
				m(this.navi)
			]),
			m("main.container", [
				m(this.main)
			])
		]
	}
}

let proxy = new ProxyWiki
let prop = new PropertyWiki(proxy)
let view = new ViewWikipage(prop)
let navi = new ViewNavigate()
let body = new ViewLayout(navi, view)

m.route(document.body, "/page/home", {
	"/page/:id": {
		onmatch: (param) => {
			if (prop.id != param.id) {
				prop = new PropertyWiki(proxy)
				view = new ViewWikipage(prop)
				body.main = view
				return m(body)
			}
		},
		render: () => {
			return m(body)
		}
	},
})
