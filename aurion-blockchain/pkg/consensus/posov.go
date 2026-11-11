package consensus

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	pb_block "github.com/elibend220/aurion-blockchain/pb/pb_block"
	pb_consensus "github.com/elibend220/aurion-blockchain/pb/pb_consensus"
	pb_transaction "github.com/elibend220/aurion-blockchain/pb/pb_transaction"
)

// ProofOfSovereigntyEngine implements the PoSov consensus mechanism
type ProofOfSovereigntyEngine struct {
	mu                sync.RWMutex
	validatorSet      *pb_consensus.ValidatorSet
	currentHeight     uint64
	currentRound      uint64
	currentPhase      Phase
	quorumSize        int
	proposedBlock     *pb_block.Block
	votes             map[string]*pb_consensus.VoteMessage
	prevotes          int
	precommits        int
	timeoutDuration   time.Duration
	lastBlockTime     time.Time
	finalizationChan  chan *pb_block.Block
}

// Phase represents the consensus phase
type Phase int

const (
	PHASE_PROPOSE Phase = iota
	PHASE_PREVOTE
	PHASE_PRECOMMIT
	PHASE_COMMIT
)

// NewProofOfSovereigntyEngine creates a new PoSov consensus engine
func NewProofOfSovereigntyEngine(validatorSet *pb_consensus.ValidatorSet) *ProofOfSovereigntyEngine {
	return &ProofOfSovereigntyEngine{
		validatorSet:    validatorSet,
		currentHeight:   1,
		currentRound:    0,
		currentPhase:    PHASE_PROPOSE,
		quorumSize:      int(validatorSet.QuorumSize),
		votes:           make(map[string]*pb_consensus.VoteMessage),
		timeoutDuration: 15 * time.Second,
		finalizationChan: make(chan *pb_block.Block, 100),
	}
}

// GetCurrentLeader returns the current round's proposer (deterministic round-robin)
func (p *ProofOfSovereigntyEngine) GetCurrentLeader() string {
	validatorCount := len(p.validatorSet.Validators)
	leaderIndex := int(p.currentRound) % validatorCount
	return p.validatorSet.Validators[leaderIndex].ValidatorId
}

// ProposeBlock is called by the leader to propose a new block
func (p *ProofOfSovereigntyEngine) ProposeBlock(block *pb_block.Block, proposerPrivateKey ed25519.PrivateKey) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Verify proposer is the current leader
	currentLeader := p.GetCurrentLeader()
	if block.Header.Height != p.currentHeight {
		return fmt.Errorf("block height %d doesn't match current height %d", block.Header.Height, p.currentHeight)
	}

	// Sign the block
	blockHash := hashBlock(block)
	signature := ed25519.Sign(proposerPrivateKey, blockHash)

	// Create proposal message
	proposal := &pb_consensus.ProposalMessage{
		Round:             p.currentRound,
		BlockHeight:       block.Header.Height,
		Block:             block,
		Proposer:          currentLeader,
		ProposerSignature: signature,
		Timestamp:         uint64(time.Now().Unix()),
	}

	p.proposedBlock = block
	p.currentPhase = PHASE_PREVOTE

	fmt.Printf("✅ Block proposed at height %d round %d by %s\n", block.Header.Height, p.currentRound, currentLeader)
	return nil
}

// ReceiveVote processes an incoming vote from a validator
func (p *ProofOfSovereigntyEngine) ReceiveVote(vote *pb_consensus.VoteMessage) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Basic validation
	if vote.Round != p.currentRound || vote.BlockHeight != p.currentHeight {
		return fmt.Errorf("vote round/height mismatch")
	}

	// Verify voter is a known validator
	if !p.isValidValidator(vote.Voter) {
		return fmt.Errorf("unknown validator: %s", vote.Voter)
	}

	// Store vote
	voteKey := fmt.Sprintf("%s_%d", vote.Voter, vote.VoteType)
	p.votes[voteKey] = vote

	// Count votes based on type
	switch vote.VoteType {
	case pb_consensus.VoteMessage_PREVOTE:
		p.prevotes++
		fmt.Printf("📊 PREVOTE received from %s (%d/%d)\n", vote.Voter, p.prevotes, p.quorumSize)

		if p.prevotes >= p.quorumSize {
			p.currentPhase = PHASE_PRECOMMIT
			fmt.Println("✅ PREVOTE threshold reached, moving to PRECOMMIT phase")
		}

	case pb_consensus.VoteMessage_PRECOMMIT:
		p.precommits++
		fmt.Printf("📊 PRECOMMIT received from %s (%d/%d)\n", vote.Voter, p.precommits, p.quorumSize)

		if p.precommits >= p.quorumSize {
			p.currentPhase = PHASE_COMMIT
			fmt.Println("✅ PRECOMMIT threshold reached, committing block!")
			p.finalizationChan <- p.proposedBlock
		}
	}

	return nil
}

// FinalizeBlock commits a block to the ledger
func (p *ProofOfSovereigntyEngine) FinalizeBlock(block *pb_block.Block, validatorSignatures [][]byte) (*pb_block.Block, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Verify we have enough signatures (3-of-4)
	if len(validatorSignatures) < p.quorumSize {
		return nil, fmt.Errorf("insufficient signatures: %d < %d", len(validatorSignatures), p.quorumSize)
	}

	// Update block header with finalization signatures
	block.Header.Signature = validatorSignatures

	p.lastBlockTime = time.Now()
	p.currentHeight++
	p.currentRound = 0
	p.currentPhase = PHASE_PROPOSE
	p.prevotes = 0
	p.precommits = 0
	p.votes = make(map[string]*pb_consensus.VoteMessage)

	fmt.Printf("✅ BLOCK FINALIZED: Height %d, Hash: %s\n", block.Header.Height, hex.EncodeToString(hashBlock(block)))

	return block, nil
}

// TransitionPhase moves to the next consensus phase
func (p *ProofOfSovereigntyEngine) TransitionPhase() {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.currentPhase {
	case PHASE_PROPOSE:
		p.currentPhase = PHASE_PREVOTE
		fmt.Println("📍 Phase: PROPOSE → PREVOTE")

	case PHASE_PREVOTE:
		if p.prevotes >= p.quorumSize {
			p.currentPhase = PHASE_PRECOMMIT
			fmt.Println("📍 Phase: PREVOTE → PRECOMMIT")
		}

	case PHASE_PRECOMMIT:
		if p.precommits >= p.quorumSize {
			p.currentPhase = PHASE_COMMIT
			fmt.Println("📍 Phase: PRECOMMIT → COMMIT")
		}

	case PHASE_COMMIT:
		// Move to next round
		p.currentRound++
		p.currentPhase = PHASE_PROPOSE
		fmt.Printf("📍 Phase: COMMIT → PROPOSE (New Round %d)\n", p.currentRound)
	}
}

// HandleTimeout skips the current proposer if timeout occurs
func (p *ProofOfSovereigntyEngine) HandleTimeout() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	currentLeader := p.GetCurrentLeader()
	fmt.Printf("⏱️  Timeout! Proposer %s didn't respond, moving to next round\n", currentLeader)

	p.currentRound++
	p.currentPhase = PHASE_PROPOSE
	p.prevotes = 0
	p.precommits = 0
	p.votes = make(map[string]*pb_consensus.VoteMessage)

	newLeader := p.GetCurrentLeader()
	fmt.Printf("✅ New leader elected: %s (Round %d)\n", newLeader, p.currentRound)

	return nil
}

// GetConsensusState returns the current consensus state
func (p *ProofOfSovereigntyEngine) GetConsensusState() *pb_consensus.ConsensusState {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var phase pb_consensus.ConsensusState_Phase
	switch p.currentPhase {
	case PHASE_PROPOSE:
		phase = pb_consensus.ConsensusState_PROPOSE
	case PHASE_PREVOTE:
		phase = pb_consensus.ConsensusState_PREVOTE
	case PHASE_PRECOMMIT:
		phase = pb_consensus.ConsensusState_PRECOMMIT
	case PHASE_COMMIT:
		phase = pb_consensus.ConsensusState_COMMIT
	}

	return &pb_consensus.ConsensusState{
		Height:        p.currentHeight,
		Round:         p.currentRound,
		Phase:         phase,
		Proposer:      p.GetCurrentLeader(),
		PrevoteCount:  int32(p.prevotes),
		PrecommitCount: int32(p.precommits),
	}
}

// GetNextLeader returns who will be the leader in N rounds
func (p *ProofOfSovereigntyEngine) GetNextLeader(roundsAhead uint64) string {
	nextRound := p.currentRound + roundsAhead
	validatorCount := len(p.validatorSet.Validators)
	leaderIndex := int(nextRound) % validatorCount
	return p.validatorSet.Validators[leaderIndex].ValidatorId
}

// isValidValidator checks if a validator ID is in the validator set
func (p *ProofOfSovereigntyEngine) isValidValidator(validatorID string) bool {
	for _, v := range p.validatorSet.Validators {
		if v.ValidatorId == validatorID {
			return true
		}
	}
	return false
}

// hashBlock creates a SHA256 hash of a block
func hashBlock(block *pb_block.Block) []byte {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%d%d%s", block.Header.Height, block.Header.Timestamp, block.Header.PreviousHash)))
	return hash.Sum(nil)
}

// ConsensusStats provides statistics about consensus
type ConsensusStats struct {
	Height          uint64
	Round           uint64
	Phase           string
	CurrentLeader   string
	Prevotes        int
	Precommits      int
	ValidatorCount  int
	Uptime          time.Duration
}

// GetStats returns consensus statistics
func (p *ProofOfSovereigntyEngine) GetStats() *ConsensusStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	phaseStr := "UNKNOWN"
	switch p.currentPhase {
	case PHASE_PROPOSE:
		phaseStr = "PROPOSE"
	case PHASE_PREVOTE:
		phaseStr = "PREVOTE"
	case PHASE_PRECOMMIT:
		phaseStr = "PRECOMMIT"
	case PHASE_COMMIT:
		phaseStr = "COMMIT"
	}

	uptime := time.Since(p.lastBlockTime)

	return &ConsensusStats{
		Height:         p.currentHeight,
		Round:          p.currentRound,
		Phase:          phaseStr,
		CurrentLeader:  p.GetCurrentLeader(),
		Prevotes:       p.prevotes,
		Precommits:     p.precommits,
		ValidatorCount: len(p.validatorSet.Validators),
		Uptime:         uptime,
	}
}

// ConsensusLoop runs the main consensus loop
func (p *ProofOfSovereigntyEngine) ConsensusLoop(blockProposalChan chan *pb_block.Block, voteChan chan *pb_consensus.VoteMessage) {
	ticker := time.NewTicker(p.timeoutDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Timeout occurred
			p.HandleTimeout()

		case vote := <-voteChan:
			// Received a vote
			if err := p.ReceiveVote(vote); err != nil {
				fmt.Printf("❌ Error receiving vote: %v\n", err)
			}

		case block := <-blockProposalChan:
			// Received a block proposal
			if p.GetCurrentLeader() == block.Header.SignerId[0] { // Simplified check
				fmt.Printf("✅ Received block proposal from current leader\n")
				p.proposedBlock = block
			}

		case block := <-p.finalizationChan:
			// Block is finalized
			fmt.Printf("✅ Block %d finalized and added to ledger\n", block.Header.Height)
			p.currentHeight++
		}
	}
}
