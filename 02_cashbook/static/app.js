"use strict"

class ProxyCashbook {
	submit(list) {
		return m.request({
			method: "POST",
			url: "//localhost:4989/api/cashbook",
			body: { entries: list }
		})
	}
}

class PropertyCashbookDetail {
	constructor() {
		this.date = ""
		this.item = ""
		this.amount = ""
	}
}

class PropertyCashbook {

	constructor(proxy) {
		this.proxy = proxy
		this.reset()
	}

	reset() {
		this.detail = {}
		this.total = 0
		this.result = ""
		this.detail[Date.now()] = new PropertyCashbookDetail()
	}

	sumTotalAmount() {
		if (Object.keys(this.detail).length == 0) {
			return 0
		}
		return Object.keys(this.detail).map((k) => {
			let amount = parseInt(this.detail[k].amount, 10)
			return isNaN(amount)? 0: amount
		}).reduce((sum,amount) => {
			return sum + amount
		})
	}

	submit() {
		let list = Object.keys(this.detail).map((k) => {
			return {
				date: this.detail[k].date,
				item: this.detail[k].item,
				amount: this.detail[k].amount
			}
		})
		proxy.submit(list).then((result) => {
			this.result = result
		})
		.catch((e) => {
			this.result = `${e.code} ${e.message}`
		})
	}
}

class ViewCashbook {

	constructor(prop) {
		this.prop = prop
	}

	view() {
		return [
			m("h3", "お買い物帳"),

			m("div", [
				m("div.row", [
					m("label.col-3.col", "日付"),
					m("label.col-4.col", "買った物"),
					m("label.col-3.col", "金額"),
					m("label.col-2.col", ""),
				]),

				...Object.keys(this.prop.detail).map((k) => {
					return m("div", [
						m("div.row", [
							m("input.col-3.col[type=date]", {
								value: this.prop.detail[k].date,
								oninput: (e) => { this.prop.detail[k].date = e.target.value }
							}),
							m("input.col-4.col[type=text]", {
								value: this.prop.detail[k].item,
								oninput: (e) => { this.prop.detail[k].item = e.target.value }
							}),
							m("input.col-3.col[type=text]", {
								value: this.prop.detail[k].amount,
								oninput: (e) => {
									this.prop.detail[k].amount = e.target.value.replace(/[^0-9]/g,'')
									this.prop.total = this.prop.sumTotalAmount()
								}
							}),
							m("label.col-2.col", [
								m("a", {
									onclick: () => {
										delete this.prop.detail[k]
										this.prop.total = this.prop.sumTotalAmount()
									}
								},"削除"),
							])
						]),
					])
				}),

				m("div.row.flex-center", [
					m("input[type=button][value=行を追加]", {
						onclick: () => {
							this.prop.detail[Date.now()] = new PropertyCashbookDetail()
						},
					}),
				]),
			]),

			m("div.row.flex-edges", [
				m("div.col", [
					m("p", `総額は ${Number(this.prop.total).toLocaleString()} 円です`),
				]),
				m("input.col.btn-success[type=button][value=記帳する]", {
					onclick: () => {
						this.prop.submit()
						m.route.set("/result")
					},
				}),
			]),
		]
	}
}

class ViewResult {

	constructor(prop) {
		this.prop = prop
	}

	view() {
		return [
			m("h3", "お買い物帳の記帳結果"),
			m("p", `${JSON.stringify(this.prop.result)}`),
		]
	}
}

class ViewLayout {
	constructor(main) {
		this.main = main
	}

	view() {
		return [
			m("main.paper.container", [
				m(this.main)
			])
		]
	}
}

const proxy = new ProxyCashbook
const prop = new PropertyCashbook(proxy)
const book = new ViewLayout(new ViewCashbook(prop))
const result = new ViewLayout(new ViewResult(prop))

m.route(document.body, "/", {
	"/": {
		render: () => {
			return m(book)
		}
	},
	"/result": {
		render: () => {
			return m(result)
		}
	},
})
