module gitlab.com/thorchain/benchmark_tss

go 1.13

require (
	github.com/binance-chain/go-sdk v1.2.1
	github.com/cosmos/cosmos-sdk v0.38.4
	github.com/libp2p/go-libp2p-core v0.5.6
	github.com/rs/zerolog v1.17.2
	github.com/tendermint/tendermint v0.33.3
	github.com/zondax/ledger-go v0.11.0 // indirect
	gitlab.com/thorchain/bepswap/thornode v0.0.0-20191121232047-8acb6f8cb031
	gitlab.com/thorchain/tss/go-tss v0.0.0-20200521211844-8c2925d834b5
)

replace github.com/binance-chain/go-sdk => gitlab.com/thorchain/binance-sdk v1.2.2
