package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/zeebo/blake3"
)

type Work struct {
	HeaderHash  common.Hash
	SeedHash    common.Hash
	Target      *big.Int
	JobID       string
	BlockNumber uint64
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./liberty-miner <rpc-url> <CPU threads>")
		return
	}

	log.Println(`


	╭╮╱╱╱╭╮╱╱╱╱╱╱╭╮╱╱╱╱╱╱╱╭━╮╭━╮
	┃┃╱╱╱┃┃╱╱╱╱╱╭╯╰╮╱╱╱╱╱╱┃┃╰╯┃┃
	┃┃╱╱╭┫╰━┳━━┳┻╮╭╋╮╱╭╮╱╱┃╭╮╭╮┣┳━╮╭━━┳━╮
	┃┃╱╭╋┫╭╮┃┃━┫╭┫┃┃┃╱┃┣━━┫┃┃┃┃┣┫╭╮┫┃━┫╭╯
	┃╰━╯┃┃╰╯┃┃━┫┃┃╰┫╰━╯┣━━┫┃┃┃┃┃┃┃┃┃┃━┫┃
	╰━━━┻┻━━┻━━┻╯╰━┻━╮╭╯╱╱╰╯╰╯╰┻┻╯╰┻━━┻╯
	╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╭━╯┃
	╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╰━━╯

		`)

	rpcURL := os.Args[1]
	var threads int
	fmt.Sscanf(os.Args[2], "%d", &threads)

	client, err := rpc.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to RPC: %v", err)
	}

	log.Println("Liberty Project: Connected to node successfully.")

	mine(client, threads)
}

func mine(client *rpc.Client, threads int) {
	var wg sync.WaitGroup
	workCh := make(chan *Work)
	var prevWork *Work

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, workCh, client)
		}(i)
	}

	for {
		work, err := getWork(client)
		if err != nil {
			log.Printf("Error fetching work: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if prevWork == nil || work.JobID != prevWork.JobID {
			log.Printf("New LBRT mining job received: Block=%d, JobID=%s", work.BlockNumber, work.JobID)

			for i := 0; i < threads; i++ {
				workCh <- work
			}
			prevWork = work
		}
		time.Sleep(1 * time.Second)
	}
}

func worker(id int, workCh <-chan *Work, client *rpc.Client) {
	rand.Seed(time.Now().UnixNano() + int64(id))

	var (
		currentCtx    context.Context
		currentCancel context.CancelFunc
	)

	for {
		currentWork := <-workCh

		if currentCancel != nil {
			currentCancel()
		}

		currentCtx, currentCancel = context.WithCancel(context.Background())
		go mineBlock(currentCtx, currentWork, id, client)
	}
}

func mineBlock(ctx context.Context, work *Work, id int, client *rpc.Client) {
	var (
		headerHash = work.HeaderHash.Bytes()
		target     = work.Target
		nonce      = uint64(rand.Int63())
		powBuffer  = new(big.Int)
		iterCount  = 312688
	)
	log.Printf("Worker %d: Starting mining with nonce=%d", id, nonce)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var buffer bytes.Buffer
			buffer.Write(headerHash)
			nonceBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(nonceBytes, nonce)
			buffer.Write(nonceBytes)

			hashResult := blake3.Sum256(buffer.Bytes())
			for i := 0; i < iterCount; i++ {
				hashResult = blake3.Sum256(hashResult[:])
			}

			powBuffer.SetBytes(hashResult[:])

			if powBuffer.Cmp(target) <= 0 {
				mixDigest := common.BytesToHash(hashResult[:])
				encodedNonce := [8]byte{}
				binary.BigEndian.PutUint64(encodedNonce[:], nonce)

				log.Printf("Worker %d: Valid solution found, submitting to node", id)
				err := submitWork(client, encodedNonce, work.HeaderHash, mixDigest, id)
				if err != nil {
					log.Printf("Worker %d: Submission failed: %v", id, err)
				} else {
					log.Printf("Worker %d: Solution accepted by the node", id)
				}
				return
			}
			nonce++
		}
	}
}

func getWork(client *rpc.Client) (*Work, error) {
	var result [3]string
	err := client.Call(&result, "eth_getWork")
	if err != nil {
		return nil, err
	}

	if len(result) < 3 {
		return nil, fmt.Errorf("Invalid response from eth_getWork")
	}

	headerHash := common.HexToHash(result[0])
	seedHash := common.HexToHash(result[1])
	targetBytes, err := hex.DecodeString(result[2][2:])
	if err != nil {
		return nil, fmt.Errorf("Failed to decode target: %v", err)
	}
	target := new(big.Int).SetBytes(targetBytes)

	var blockNumberHex string
	err = client.Call(&blockNumberHex, "eth_blockNumber")
	if err != nil {
		return nil, fmt.Errorf("Failed to get block number: %v", err)
	}
	blockNumber, err := strconv.ParseUint(blockNumberHex[2:], 16, 64)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse block number: %v", err)
	}

	return &Work{
		HeaderHash:  headerHash,
		SeedHash:    seedHash,
		Target:      target,
		JobID:       headerHash.Hex(),
		BlockNumber: blockNumber,
	}, nil
}

func submitWork(client *rpc.Client, nonce [8]byte, headerHash common.Hash, mixDigest common.Hash, id int) error {
	log.Printf("Worker %d: Submitting solution", id)

	var result bool
	err := client.Call(&result, "eth_submitWork", hexutil.Encode(nonce[:]), headerHash.Hex(), mixDigest.Hex())
	if err != nil {
		return fmt.Errorf("Submission failed: %w", err)
	}

	if !result {
		return fmt.Errorf("Solution rejected by the node")
	}
	return nil
}
