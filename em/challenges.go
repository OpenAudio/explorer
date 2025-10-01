package em

// ChallengeID represents the challenge identifier enum
type ChallengeID string

const (
	ChallengeTrackUploads                ChallengeID = "u"
	ChallengeReferrals                   ChallengeID = "r"
	ChallengeVerifiedReferrals           ChallengeID = "rv"
	ChallengeReferred                    ChallengeID = "rd"
	ChallengeMobileInstall               ChallengeID = "m"
	ChallengeConnectVerifiedAccount      ChallengeID = "v"
	ChallengeListenStreakEndless         ChallengeID = "e"
	ChallengeCompleteProfile             ChallengeID = "p"
	ChallengeSendFirstTip                ChallengeID = "ft"
	ChallengeCreateFirstPlaylist         ChallengeID = "fp"
	ChallengeAudioMatchingBuyer          ChallengeID = "b"
	ChallengeAudioMatchingSeller         ChallengeID = "s"
	ChallengeTrendingTrack               ChallengeID = "tt"
	ChallengeTrendingPlaylist            ChallengeID = "tp"
	ChallengeTrendingUndergroundTrack    ChallengeID = "tut"
	ChallengeOneShot                     ChallengeID = "o"
	ChallengeFirstWeeklyComment          ChallengeID = "c"
	ChallengePlayCount250Milestone2025   ChallengeID = "p1"
	ChallengePlayCount1000Milestone2025  ChallengeID = "p2"
	ChallengePlayCount10000Milestone2025 ChallengeID = "p3"
	ChallengeTastemaker                  ChallengeID = "t"
	ChallengeCosign                      ChallengeID = "cs"
	ChallengePinnedComment               ChallengeID = "cp"
	ChallengeRemixContestWinner          ChallengeID = "w"
)

// DefaultSpecifier represents the default challenge specifier
type DefaultSpecifier struct {
	ChallengeID ChallengeID `json:"challengeId"`
	UserID      string      `json:"userId"`
}

// ReferralSpecifier represents a referral challenge specifier
type ReferralSpecifier struct {
	ChallengeID    ChallengeID `json:"challengeId"`
	UserID         string      `json:"userId"`
	ReferredUserID string      `json:"referredUserId"`
}

// AudioMatchSpecifier represents an audio matching challenge specifier
type AudioMatchSpecifier struct {
	ChallengeID ChallengeID `json:"challengeId"`
	UserID      string      `json:"userId"`
	ContentID   string      `json:"contentId"`
}

// GenerateSpecifierRequest can be one of the specifier types
type GenerateSpecifierRequest interface{}

// ClaimRewardsRequest represents a request to claim challenge rewards
type ClaimRewardsRequest struct {
	ChallengeID ChallengeID `json:"challengeId"`
	Specifier   string      `json:"specifier"`
	Amount      interface{} `json:"amount"` // bigint or number (wAUDIO or wAUDIO Wei)
	UserID      string      `json:"userId"`
}

// AttestationTransactionSignature represents an attestation transaction signature
type AttestationTransactionSignature struct {
	TransactionSignature      string `json:"transactionSignature"`
	AntiAbuseOracleEthAddress string `json:"antiAbuseOracleEthAddress"`
}

// AAOErrorResponse represents an AAO error response
type AAOErrorResponse struct {
	AAOErrorCode int `json:"aaoErrorCode"`
}

// ClaimAllRewardsRequest represents a request to claim all rewards
type ClaimAllRewardsRequest struct {
	UserID      *string      `json:"userId,omitempty"`
	ChallengeID *ChallengeID `json:"challengeId,omitempty"`
	Specifier   *string      `json:"specifier,omitempty"`
}

// ClaimResult represents the result of a claim operation
type ClaimResult struct {
	ChallengeID string   `json:"challengeId"`
	Specifier   string   `json:"specifier"`
	Amount      string   `json:"amount"`
	Signatures  []string `json:"signatures"`
	Error       *string  `json:"error,omitempty"`
}

// ClaimAllResponseBody represents the response body for claiming all rewards
type ClaimAllResponseBody struct {
	Data []ClaimResult `json:"data"`
}
