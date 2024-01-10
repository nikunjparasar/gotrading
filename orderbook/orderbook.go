package orderbook

import (
	"fmt"
	"sort"
	"time"
)

/*
-$$$$$$\  $$$$$$$\  $$$$$$$\  $$$$$$$$\ $$$$$$$\
$$  __$$\ $$  __$$\ $$  __$$\ $$  _____|$$  __$$\
$$ /  $$ |$$ |  $$ |$$ |  $$ |$$ |      $$ |  $$ |
$$ |  $$ |$$$$$$$  |$$ |  $$ |$$$$$\    $$$$$$$  |
$$ |  $$ |$$  __$$< $$ |  $$ |$$  __|   $$  __$$<
$$ |  $$ |$$ |  $$ |$$ |  $$ |$$ |      $$ |  $$ |
-$$$$$$  |$$ |  $$ |$$$$$$$  |$$$$$$$$\ $$ |  $$ |
\______/ \__|  \__|\_______/ \________|\__|  \__|
*/
const (
	BUY  = "BUY"
	SELL = "SELL"
)

type Order struct {
	ordersize float64
	action    string
	lim       *Limit
	timestamp int64
}

func (o *Order) Limit() *Limit    { return o.lim }
func (o *Order) Action() string   { return o.action }
func (o *Order) Size() float64    { return o.ordersize }
func (o *Order) Timestamp() int64 { return o.timestamp }

type Match struct {
	ask        *Order
	bid        *Order
	sizefilled float64
	price      float64
}

// Constructor for Order type
func NewOrder(ac string, size float64) *Order {
	return &Order{
		ordersize: size,
		action:    ac,
		timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f]", o.ordersize)
}

func (o *Order) IsFilled() bool {
	return o.ordersize == 0
}

// for sorting orders by timestamp
type Orders []*Order

func (o Orders) Len() int { return len(o) }
func (o Orders) Swap(i, j int) {
	temp := o[i]
	o[i] = o[j]
	o[j] = temp
}
func (o Orders) Less(i, j int) bool { return o[i].timestamp < o[j].timestamp }

/*
$$\       $$$$$$\ $$\      $$\ $$$$$$\ $$$$$$$$\
$$ |      \_$$  _|$$$\    $$$ |\_$$  _|\__$$  __|
$$ |        $$ |  $$$$\  $$$$ |  $$ |     $$ |
$$ |        $$ |  $$\$$\$$ $$ |  $$ |     $$ |
$$ |        $$ |  $$ \$$$  $$ |  $$ |     $$ |
$$ |        $$ |  $$ |\$  /$$ |  $$ |     $$ |
$$$$$$$$\ $$$$$$\ $$ | \_/ $$ |$$$$$$\    $$ |
\________|\______|\__|     \__|\______|   \__|



*/

type Limit struct {
	price    float64
	orders   Orders
	totalvol float64
}

// constructor for a Limit
func NewLimit(setprice float64) *Limit {
	return &Limit{
		price:  setprice,
		orders: []*Order{},
	}
}

func (l *Limit) Orders() Orders       { return l.orders }
func (l *Limit) Price() float64       { return l.price }
func (l *Limit) TotalVolume() float64 { return l.totalvol }

// add an order to a limit by appending to the orders slice
func (l *Limit) AddOrder(o *Order) {
	o.lim = l
	l.orders = append(l.orders, o)
	l.totalvol += o.ordersize
}

func (l *Limit) DeleteOrder(o *Order) {
	for i := 0; i < len(l.orders); i++ {
		if l.orders[i] == o {
			// delete order efficiently by moving to end of slice and slicing off
			l.orders[i] = l.orders[len(l.orders)-1]
			l.orders = l.orders[:len(l.orders)-1]
		}
	}

	o.lim = nil
	l.totalvol -= o.ordersize

	sort.Sort(l.orders) //////////////////////////////////////////////////////////////////////// OPTIMIZE
}

func (l *Limit) fillOrder(a, b *Order) Match {
	var (
		buy      *Order
		sell     *Order
		fillsize float64
	)

	// check which order is the bid vs ask
	if a.action == BUY {
		buy = a
		sell = b
	} else {
		buy = b
		sell = a
	}

	// fill the max possible order
	if buy.ordersize >= sell.ordersize {
		fillsize = sell.ordersize
		buy.ordersize -= sell.ordersize
		sell.ordersize = 0
	} else {
		fillsize = buy.ordersize
		sell.ordersize -= buy.ordersize
		buy.ordersize = 0
	}

	return Match{
		ask:        sell,
		bid:        buy,
		sizefilled: fillsize,
		price:      l.price,
	}

}

func (l *Limit) Fill(o *Order) []Match {
	var (
		matches        []Match
		ordersToDelete []*Order
	)

	for _, order := range l.orders {
		match := l.fillOrder(order, o)
		matches = append(matches, match)

		l.totalvol -= match.sizefilled

		if order.IsFilled() {
			ordersToDelete = append(ordersToDelete, order)
		}

		if o.IsFilled() {
			break
		}
	}

	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}

	return matches
}

/*

 $$$$$$\  $$$$$$$\  $$$$$$$\  $$$$$$$$\ $$$$$$$\  $$$$$$$\   $$$$$$\   $$$$$$\  $$\   $$\
$$  __$$\ $$  __$$\ $$  __$$\ $$  _____|$$  __$$\ $$  __$$\ $$  __$$\ $$  __$$\ $$ | $$  |
$$ /  $$ |$$ |  $$ |$$ |  $$ |$$ |      $$ |  $$ |$$ |  $$ |$$ /  $$ |$$ /  $$ |$$ |$$  /
$$ |  $$ |$$$$$$$  |$$ |  $$ |$$$$$\    $$$$$$$  |$$$$$$$\ |$$ |  $$ |$$ |  $$ |$$$$$  /
$$ |  $$ |$$  __$$< $$ |  $$ |$$  __|   $$  __$$< $$  __$$\ $$ |  $$ |$$ |  $$ |$$  $$<
$$ |  $$ |$$ |  $$ |$$ |  $$ |$$ |      $$ |  $$ |$$ |  $$ |$$ |  $$ |$$ |  $$ |$$ |\$$\
 $$$$$$  |$$ |  $$ |$$$$$$$  |$$$$$$$$\ $$ |  $$ |$$$$$$$  | $$$$$$  | $$$$$$  |$$ | \$$\
 \______/ \__|  \__|\_______/ \________|\__|  \__|\_______/  \______/  \______/ \__|  \__|
*/

type Orderbook struct {
	asks []*Limit
	bids []*Limit

	askLimits map[float64]*Limit
	bidLimits map[float64]*Limit
}

func (ob *Orderbook) BidsTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.bids); i++ {
		totalVolume += ob.bids[i].totalvol
	}

	return totalVolume
}

func (ob *Orderbook) AsksTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.asks); i++ {
		totalVolume += ob.asks[i].totalvol
	}

	return totalVolume
}

func (ob *Orderbook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks}) //////////////////////////////////////////////////////////////// OPTIMIZE
	return ob.asks
}

func (ob *Orderbook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids}) //////////////////////////////////////////////////////////////// OPTIMIZE
	return ob.bids
}

// orderbook constructor
func NewOrderbook() *Orderbook {
	return &Orderbook{
		asks:      []*Limit{},
		bids:      []*Limit{},
		askLimits: make(map[float64]*Limit),
		bidLimits: make(map[float64]*Limit),
	}
}

func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {
	matches := []Match{}

	if o.action == BUY {
		if o.ordersize > ob.AsksTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] to fill market order [size: %.2f]", ob.AsksTotalVolume(), o.ordersize))
		}

		for _, limit := range ob.Asks() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.orders) == 0 {
				ob.clearLimits(SELL, limit)
			}
		}
	} else {
		if o.ordersize > ob.BidsTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] to fill market order [size: %.2f]", ob.AsksTotalVolume(), o.ordersize))
		}

		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.orders) == 0 {
				ob.clearLimits(BUY, limit)
			}
		}
	}

	return matches
}

func (ob *Orderbook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

	if o.action == BUY {
		limit = ob.bidLimits[price]
	} else {
		limit = ob.askLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)

		if o.action == BUY {
			ob.bids = append(ob.bids, limit)
			ob.bidLimits[price] = limit
		} else {
			ob.asks = append(ob.asks, limit)
			ob.askLimits[price] = limit
		}
	}
	limit.AddOrder(o)

}

func (ob *Orderbook) CancelOrder(o *Order) {
	limit := o.lim
	limit.DeleteOrder(o)
}

func (ob *Orderbook) clearLimits(action string, l *Limit) {
	if action == BUY {
		delete(ob.bidLimits, l.price)
		for i := 0; i < len(ob.bids); i++ {
			if ob.bids[i] == l {
				ob.bids[i] = ob.bids[len(ob.bids)-1]
				ob.bids = ob.bids[:len(ob.bids)-1]
			}

		}
	} else {
		delete(ob.askLimits, l.price)
		for i := 0; i < len(ob.asks); i++ {
			if ob.asks[i] == l {
				ob.asks[i] = ob.asks[len(ob.asks)-1]
				ob.asks = ob.asks[:len(ob.asks)-1]
			}

		}
	}
}

/*
 $$$$$$\  $$$$$$$\ $$$$$$$$\ $$$$$$\ $$\      $$\ $$$$$$\ $$$$$$$$\  $$$$$$\ $$$$$$$$\ $$$$$$\  $$$$$$\  $$\   $$\
$$  __$$\ $$  __$$\\__$$  __|\_$$  _|$$$\    $$$ |\_$$  _|\____$$  |$$  __$$\\__$$  __|\_$$  _|$$  __$$\ $$$\  $$ |
$$ /  $$ |$$ |  $$ |  $$ |     $$ |  $$$$\  $$$$ |  $$ |      $$  / $$ /  $$ |  $$ |     $$ |  $$ /  $$ |$$$$\ $$ |
$$ |  $$ |$$$$$$$  |  $$ |     $$ |  $$\$$\$$ $$ |  $$ |     $$  /  $$$$$$$$ |  $$ |     $$ |  $$ |  $$ |$$ $$\$$ |
$$ |  $$ |$$  ____/   $$ |     $$ |  $$ \$$$  $$ |  $$ |    $$  /   $$  __$$ |  $$ |     $$ |  $$ |  $$ |$$ \$$$$ |
$$ |  $$ |$$ |        $$ |     $$ |  $$ |\$  /$$ |  $$ |   $$  /    $$ |  $$ |  $$ |     $$ |  $$ |  $$ |$$ |\$$$ |
 $$$$$$  |$$ |        $$ |   $$$$$$\ $$ | \_/ $$ |$$$$$$\ $$$$$$$$\ $$ |  $$ |  $$ |   $$$$$$\  $$$$$$  |$$ | \$$ |
 \______/ \__|        \__|   \______|\__|     \__|\______|\________|\__|  \__|  \__|   \______| \______/ \__|  \__|

*/

// for sorting limits by best ask or best bid (OLD WAY) takes O(NLOGN) each time
type Limits []*Limit

// ByBestAst is a type for sorting limits by best ask
type ByBestAsk struct{ Limits }

func (a ByBestAsk) Len() int { return len(a.Limits) }
func (a ByBestAsk) Swap(i, j int) {
	temp := a.Limits[i]
	a.Limits[i] = a.Limits[j]
	a.Limits[j] = temp
}
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].price < a.Limits[j].price }

// ByBestBid is a type for sorting limits by best bid
type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int { return len(b.Limits) }
func (b ByBestBid) Swap(i, j int) {
	temp := b.Limits[i]
	b.Limits[i] = b.Limits[j]
	b.Limits[j] = temp
}
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].price > b.Limits[j].price }

// OPTIMIZATION REDUCE RESTRUCTURE TIME FROM O(NLOGN) TO O(LOGN) BY USING A HEAP INSTEAD OF SORTING

// type AsksHeap struct {
// 	limits []*Limit
// 	index  map[*Limit]int
// }

// // a min Heap

// func (a AsksHeap) Len() int           { return len(a.limits) }
// func (a AsksHeap) BestAsk() *Limit    { return a.limits[0] }
// func (a AsksHeap) Less(i, j int) bool { return a.limits[i].Price < a.limits[j].Price }
// func (a AsksHeap) Insert(l *Limit) {
// 	// add to heap
// 	a.limits = append(a.limits, l)
// 	i := len(a.limits) // use +1 indexing
// 	//restructure heap
// 	for i > 0 && a.Less(i, i/2) {
// 		//swap elements
// 		temp := a.limits[i]
// 		a.limits[i] = a.limits[i/2]
// 		a.limits[i/2] = temp

// 		// update index mapping
// 		a.index[a.limits[i]] = i
// 		a.index[a.limits[i/2]] = i / 2

// 		i = i / 2
// 	}
// }

// func (a AsksHeap) Delete(l *Limit) {
// 	// find index of limit
// 	i := a.index[l]

// 	// swap with last element
// 	a.limits[i] = a.limits[len(a.limits)-1]
// 	a.limits = a.limits[:len(a.limits)-1]

// 	// update index mapping
// 	a.index[a.limits[i]] = i

// 	// restructure heap
// 	for i > 0 && a.Less(i, i/2) {
// 		//swap elements
// 		temp := a.limits[i]
// 		a.limits[i] = a.limits[i/2]
// 		a.limits[i/2] = temp

// 		// update index mapping
// 		a.index[a.limits[i]] = i
// 		a.index[a.limits[i/2]] = i / 2

// 		i = i / 2
// 	}
// }

// type BidsHeap []*Limit                // a max Heap
// func (b BidsHeap) Len() int           { return len(b) }
// func (a BidsHeap) BestBid() *Limit    { return a[0] }
// func (b BidsHeap) Less(i, j int) bool { return b[i].Price > b[j].Price }

// 2. use a heap to store orders

// for sorting orders by timestamp
// type Orders []*Order

// func (o Orders) Len() int { return len(o) }
// func (o Orders) Swap(i, j int) {
// 	temp := o[i]
// 	o[i] = o[j]
// 	o[j] = temp
// }
// func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }
