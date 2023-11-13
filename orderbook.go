package main

import (
	"time"
	"fmt"
)



/*
--------- ORDER ---------
*/

type Order struct {
	Size float64
	Bid bool
	Limit *Limit
	Timestamp int64
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size: size,
		Bid: bid,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f]", o.Size)
}

// for sorting orders by timestamp
type Orders []*Order

func (o Orders) Len() int { return len(o)}
func (o Orders) Swap(i, j int) { 
	temp := o[i]
	o[i] = o[j]
	o[j] = temp
}
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }



/*
--------- LIMIT ---------
*/

type Limit struct {
	Price float64
	Orders Orders
	TotalVolume float64
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price: price,
		Orders: []*Order{},
	}
}

func (l *Limit) AddOrder(o *Order){
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}


func (l *Limit) DeleteOrder(o *Order){
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

// for sorting limits by best ask or best bid

type Limits []*Limit

type ByBestAsk struct { Limits }	

func (a ByBestAsk) Len() int { return len(a.Limits)}
func (a ByBestAsk) Swap(i, j int) { 
	temp := a.Limits[i]
	a.Limits[i] = a.Limits[j]
	a.Limits[j] = temp
}
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }


type ByBestBid struct { Limits }	

func (b ByBestBid) Len() int { return len(b.Limits)}
func (b ByBestBid) Swap(i, j int) { 
	temp := b.Limits[i]
	b.Limits[i] = b.Limits[j]
	b.Limits[j] = temp
}
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }



/*
--------- MATCH ---------
*/
type Match struct {
	Ask *Order
	Bid *Order
	SizeFilled float64
	Price float64

}

/*
--------- ORDERBOOK ---------
*/


type Orderbook struct {
	Asks []*Limit
	Bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}


func NewOrderbook() *Orderbook {
	return &Orderbook{
		Asks: []*Limit{},
		Bids: []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}

func (ob *Orderbook) PlaceOrder(price float64, o *Order) []Match {
	// 1. Try to match bid or ask
		// matching logic
	// 2. Add the remaining orders to the orderbook
	

	if o.Size > 0.0 {
		ob.add(price, o)
	}

	return []Match{}
}

func (ob *Orderbook)  add(price float64, o *Order) {
	var limit *Limit

	if(o.Bid) {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if(limit == nil) {
		limit = NewLimit(price)
		limit.AddOrder(o)

		if(o.Bid) {
			ob.Bids = append(ob.Bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.Asks = append(ob.Asks, limit)	
			ob.AskLimits[price] = limit
		}
	}
}