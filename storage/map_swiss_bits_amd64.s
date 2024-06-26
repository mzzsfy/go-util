//nolint
//go:build amd64

#include "textflag.h"

// func matchMetadata(metadata *[16]int8, hash int8) uint16
// Requires: SSE2, SSSE3
TEXT ·matchMetadata(SB), NOSPLIT, $0-18
	MOVQ     metadata+0(FP), AX
	MOVBLSX  hash+8(FP), CX
	MOVD     CX, X0
	PXOR     X1, X1
	PSHUFB   X1, X0
	MOVOU    (AX), X1
	PCMPEQB  X1, X0
	PMOVMSKB X0, AX
	MOVW     AX, ret+16(FP)
	RET

TEXT ·Cpuid(SB), NOSPLIT, $0-24
	MOVL eaxArg+0(FP), AX
	MOVL ecxArg+4(FP), CX
	CPUID
	MOVL AX, eax+8(FP)
	MOVL BX, ebx+12(FP)
	MOVL CX, ecx+16(FP)
	MOVL DX, edx+20(FP)
	RET
