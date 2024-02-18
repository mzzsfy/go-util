#include "textflag.h"

TEXT ·getG(SB), NOSPLIT, $0
    MOVQ    (TLS), AX
    MOVQ    AX, ret+0(FP)
    RET
