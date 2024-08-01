"use strict"

class ProxyWords {

	loadList() {
		return m.request({
			method: "GET",
			url: "//localhost:4989/api/word",
		})
	}

	saveItem(item) {
		return m.request({
			method: "POST",
			url: "//localhost:4989/api/word",
			body: {
				word: item
			},
		})
	}
}

class PropertyWords {

	constructor(proxy) {
		this.proxy = proxy
		this.list = []
		this.item = ""
	}

	load() {
		proxy.loadList().then((result) => {
			this.list = result
		})
	}

	save() {
		proxy.saveItem(this.item).then((result) => {
			this.load()
		})
	}
}

class ViewWords {

	constructor(prop) {
		this.prop = prop
	}

	oninit() {
		this.prop.load()
	}

	view() {
		return [
			m("table", [
				m("thead", [
					m("tr", [
						m("th", "id"),
						m("th", "word")
					])
				]),
				m("tbody", this.prop.list.map((item) => {
					return m("tr", [m("td", `${item.id}`), m("td", `${item.word}`)])
				})),
			]),
			m("form", {
				onsubmit: (e) => {
					e.preventDefault()
					document.getElementById("new").value = ""
					this.prop.save()
				}
			}, [
				m("fieldset[role=group]", [
					m("input[id=new][type=text][placeholder=enter new word]", {
						oninput: (e) => {
							this.prop.item = e.target.value
						}
					}),
					m("input[type=submit]", "save"),
				]),
			]),
		]
	}
}

class ViewNavi {
	view() {
		return m("ul", [
			m("li", [
				m("strong.bold", "WORDS"),
			])
		])
	}
}

class ViewLayout {

	constructor(navi, main) {
		this.navi = navi
		this.main = main
	}

	view(vnode) {
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

const proxy = new ProxyWords
const prop = new PropertyWords(proxy)
const main = new ViewWords(prop)
const navi = new ViewNavi()
const layout = new ViewLayout(navi, main)

m.route(document.body, "/", {
	"/": {
		render: () => {
			return m(layout)
		}
	},
})
