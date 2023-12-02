package orderbook

import (
	"fmt"
	"sort"
	"time"
)

/*
$$$$$$\  $$$$$$$\  $$$$$$$\  $$$$$$$$\ $$$$$$$\
$$  __$$\ $$  __$$\ $$  __$$\ $$  _____|$$  __$$\
$$ /  $$ |$$ |  $$ |$$ |  $$ |$$ |      $$ |  $$ |
$$ |  $$ |$$$$$$$  |$$ |  $$ |$$$$$\    $$$$$$$  |
$$ |  $$ |$$  __$$< $$ |  $$ |$$  __|   $$  __$$<
$$ |  $$ |$$ |  $$ |$$ |  $$ |$$ |      $$ |  $$ |
$$$$$$  |$$ |  $$ |$$$$$$$  |$$$$$$$$\ $$ |  $$ |
\______/ \__|  \__|\_______/ \________|\__|  \__|
*/
const (
	BUY  = true
	SELL = false
)

type Order struct {
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f]", o.Size)
}

func (o *Order) IsFilled() bool {
	return o.Size == 0
}

// for sorting orders by timestamp
type Orders []*Order

func (o Orders) Len() int { return len(o) }
func (o Orders) Swap(i, j int) {
	temp := o[i]
	o[i] = o[j]
	o[j] = temp
}
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }

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
	Price       float64
	Orders      Orders
	TotalVolume float64
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

func (l *Limit) AddOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

func (l *Limit) DeleteOrder(o *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			l.Orders[i] = l.Orders[len(l.Orders)-1]
			l.Orders = l.Orders[:len(l.Orders)-1]
		}
	}

	o.Limit = nil
	l.TotalVolume -= o.Size

	sort.Sort(l.Orders)
}

func (l *Limit) fillOrder(a, b *Order) Match {
	var (
		bid        *Order
		ask        *Order
		sizeFilled float64
	)

	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}

	if bid.Size >= ask.Size {
		sizeFilled = ask.Size
		bid.Size -= ask.Size
		ask.Size = 0
	} else {
		sizeFilled = bid.Size
		ask.Size -= bid.Size
		bid.Size = 0
	}

	return Match{
		Ask:        ask,
		Bid:        bid,
		SizeFilled: sizeFilled,
		Price:      l.Price,
	}

}

func (l *Limit) Fill(o *Order) []Match {
	var (
		matches        []Match
		ordersToDelete []*Order
	)

	for _, order := range l.Orders {
		match := l.fillOrder(order, o)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

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

// for sorting limits by best ask or best bid

type Limits []*Limit

type ByBestAsk struct{ Limits }

func (a ByBestAsk) Len() int { return len(a.Limits) }
func (a ByBestAsk) Swap(i, j int) {
	temp := a.Limits[i]
	a.Limits[i] = a.Limits[j]
	a.Limits[j] = temp
}
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }

type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int { return len(b.Limits) }
func (b ByBestBid) Swap(i, j int) {
	temp := b.Limits[i]
	b.Limits[i] = b.Limits[j]
	b.Limits[j] = temp
}
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }

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

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		asks:      []*Limit{},
		bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}
func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {
	matches := []Match{}

	if o.Bid {
		if o.Size > ob.AsksTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] to fill market order [size: %.2f]", ob.AsksTotalVolume(), o.Size))
		}

		for _, limit := range ob.Asks() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimits(SELL, limit)
			}
		}
	} else {
		if o.Size > ob.BidsTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] to fill market order [size: %.2f]", ob.AsksTotalVolume(), o.Size))
		}

		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimits(BUY, limit)
			}
		}
	}

	return matches
}

func (ob *Orderbook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)

		if o.Bid {
			ob.bids = append(ob.bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.asks = append(ob.asks, limit)
			ob.AskLimits[price] = limit
		}
	}
	limit.AddOrder(o)

}

func (ob *Orderbook) CancelOrder(o *Order) {
	limit := o.Limit
	limit.DeleteOrder(o)

	// if len(limit.Orders) == 0 {
	// 	if o.Bid {
	// 		ob.clearLimits(BUY, limit)
	// 	} else {
	// 		ob.clearLimits(SELL, limit)
	// 	}
	// }
}

func (ob *Orderbook) clearLimits(bid bool, l *Limit) {
	if bid {
		delete(ob.BidLimits, l.Price)
		for i := 0; i < len(ob.bids); i++ {
			if ob.bids[i] == l {
				ob.bids[i] = ob.bids[len(ob.bids)-1]
				ob.bids = ob.bids[:len(ob.bids)-1]
			}

		}
	} else {
		delete(ob.AskLimits, l.Price)
		for i := 0; i < len(ob.asks); i++ {
			if ob.asks[i] == l {
				ob.asks[i] = ob.asks[len(ob.asks)-1]
				ob.asks = ob.asks[:len(ob.asks)-1]
			}

		}
	}
}

func (ob *Orderbook) BidsTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.bids); i++ {
		totalVolume += ob.bids[i].TotalVolume
	}

	return totalVolume
}

func (ob *Orderbook) AsksTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.asks); i++ {
		totalVolume += ob.asks[i].TotalVolume
	}

	return totalVolume
}

func (ob *Orderbook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

func (ob *Orderbook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
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

// OPTIMIZATION REDUCE RESTRUCTURE TIME FROM O(NLOGN) TO O(LOGN)

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
