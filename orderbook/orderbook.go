package orderbook

import (
	"fmt"
	"math/rand"
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
	ID        int64
	ordersize float64
	action    string
	lim       *Limit
	timestamp int64
}

// getters
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
		ID:        int64(rand.Intn(1000000000)),
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
		price:    setprice,
		orders:   []*Order{},
		totalvol: 0,
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

	sort.Sort(l.orders)
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
	asks AsksHeap
	bids BidsHeap

	askLimits map[float64]*Limit
	bidLimits map[float64]*Limit
}

// orderbook constructor
func NewOrderbook() *Orderbook {
	return &Orderbook{
		asks:      *NewAsksHeap(),
		bids:      *NewBidsHeap(),
		askLimits: make(map[float64]*Limit),
		bidLimits: make(map[float64]*Limit),
	}
}

func (ob *Orderbook) BidsTotalVolume() float64 {
	totalVol := 0.0
	for _, limit := range ob.bids.Limits {
		totalVol += limit.TotalVolume()
	}
	return totalVol
}

func (ob *Orderbook) AsksTotalVolume() float64 {
	totalVol := 0.0
	for _, limit := range ob.asks.Limits {
		totalVol += limit.TotalVolume()
	}
	return totalVol
}

func (ob *Orderbook) Asks() AsksHeap { return ob.asks }

func (ob *Orderbook) Bids() BidsHeap { return ob.bids }

func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {
	matches := []Match{}

	if o.action == BUY {
		if o.ordersize > ob.AsksTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] to fill market order [size: %.2f]", ob.AsksTotalVolume(), o.ordersize))
		}

		//get the best asks
		filledQty := 0.0
		fillSize := o.ordersize
		for filledQty < fillSize {
			bestAsk := ob.asks.Top() // Limit containing the best ask orders
			oldLimVol := bestAsk.TotalVolume()
			matches = append(matches, bestAsk.Fill(o)...)
			newLimVol := bestAsk.TotalVolume()
			filledQty += oldLimVol - newLimVol
			if bestAsk.TotalVolume() == 0 {
				ob.asks.Pop()
				delete(ob.askLimits, bestAsk.Price())
			}
		}

	} else {
		if o.ordersize > ob.BidsTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] to fill market order [size: %.2f]", ob.AsksTotalVolume(), o.ordersize))
		}
		filledQty := 0.0
		fillSize := o.ordersize
		for filledQty < fillSize {
			bestBid := ob.bids.Top() // Limit containing the best bid orders
			oldLimVol := bestBid.TotalVolume()
			matches = append(matches, bestBid.Fill(o)...)
			newLimVol := bestBid.TotalVolume()
			filledQty += oldLimVol - newLimVol
			if bestBid.TotalVolume() == 0 {
				ob.bids.Pop()
				delete(ob.bidLimits, bestBid.Price())
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
			ob.bids.Push(limit)
			ob.bidLimits[price] = limit
		} else {
			ob.asks.Push(limit)
			ob.askLimits[price] = limit
		}
	}
	limit.AddOrder(o)

}

func (ob *Orderbook) CancelOrder(o *Order) {
	limit := o.lim
	limit.DeleteOrder(o)
}

// HEAP OPTIMIZATIONS FOR ORDERBOOK LIMITS
// Reduced each market order to average O(log(n)) time complexity from O(nlog(n))
// to show the orderbook on the frontend, we can just iterate through the top limits
// will be klogn for showing it on the frontend where k is the number of limits to show

// min heap for storing asks and keeping track of total volume
type AsksHeap struct {
	Limits []*Limit
}

func NewAsksHeap() *AsksHeap {
	return &AsksHeap{
		Limits: []*Limit{},
	}
}

func (a *AsksHeap) Len() int { return len(a.Limits) }

func (a *AsksHeap) Top() *Limit {
	if len(a.Limits) == 0 {
		return nil
	}
	return a.Limits[0]
}

func (a *AsksHeap) Push(l *Limit) {
	a.Limits = append(a.Limits, l)
	a.siftup()
}

func (a *AsksHeap) Pop() *Limit {
	n := len(a.Limits)
	a.swap(0, n-1)
	val := a.Limits[n-1]
	a.siftdown(0, n-2)
	a.Limits = a.Limits[:n-1]
	return val
}

func (a *AsksHeap) swap(i, j int) {
	temp := a.Limits[i]
	a.Limits[i] = a.Limits[j]
	a.Limits[j] = temp
}

func (a *AsksHeap) siftup() {
	curridx := len(a.Limits) - 1
	parentidx := (curridx - 1) / 2
	for curridx > 0 && a.Limits[curridx].price < a.Limits[parentidx].price {
		a.swap(curridx, parentidx)
		curridx = parentidx
		parentidx = (curridx - 1) / 2
	}
}

func (a *AsksHeap) siftdown(curridx, endidx int) {
	leftidx := 2*curridx + 1
	for leftidx <= endidx {
		rightidx := curridx*2 + 2
		if rightidx > endidx {
			rightidx = -1
		}
		idxtoswap := leftidx
		if rightidx != -1 && a.Limits[rightidx].price < a.Limits[leftidx].price {
			idxtoswap = rightidx
		}

		if a.Limits[idxtoswap].price < a.Limits[curridx].price {
			a.swap(idxtoswap, curridx)
			curridx = idxtoswap
			leftidx = 2*curridx + 1
		} else {
			return
		}
	}
}

// max heap for storing bids
type BidsHeap struct {
	Limits []*Limit
}

func NewBidsHeap() *BidsHeap {
	return &BidsHeap{
		Limits: []*Limit{},
	}
}

func (b *BidsHeap) Len() int { return len(b.Limits) }
func (b *BidsHeap) swap(i, j int) {
	temp := b.Limits[i]
	b.Limits[i] = b.Limits[j]
	b.Limits[j] = temp
}

func (b *BidsHeap) Top() *Limit {
	if len(b.Limits) == 0 {
		return nil
	}
	return b.Limits[0]
}

func (b *BidsHeap) Push(l *Limit) {
	b.Limits = append(b.Limits, l)
	b.siftup()
}

func (b *BidsHeap) Pop() *Limit {
	n := len(b.Limits)
	b.swap(0, n-1)
	val := b.Limits[n-1]
	b.siftdown(0, n-2)
	b.Limits = b.Limits[:n-1]
	return val
}

func (b *BidsHeap) siftup() {
	curridx := len(b.Limits) - 1
	parentidx := (curridx - 1) / 2
	for curridx > 0 && b.Limits[curridx].price > b.Limits[parentidx].price {
		b.swap(curridx, parentidx)
		curridx = parentidx
		parentidx = (curridx - 1) / 2
	}
}

func (b *BidsHeap) siftdown(curridx, endidx int) {
	leftidx := 2*curridx + 1
	for leftidx <= endidx {
		rightidx := curridx*2 + 2
		if rightidx > endidx {
			rightidx = -1
		}
		idxtoswap := leftidx
		if rightidx != -1 && b.Limits[rightidx].price > b.Limits[leftidx].price {
			idxtoswap = rightidx
		}

		if b.Limits[idxtoswap].price > b.Limits[curridx].price {
			b.swap(idxtoswap, curridx)
			curridx = idxtoswap
			leftidx = 2*curridx + 1
		} else {
			return
		}
	}
}
