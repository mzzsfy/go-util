#include "textflag.h"

TEXT ·getG(SB), NOSPLIT, $0
    MOVL    (TLS), AX
    MOVL    AX, ret+0(FP)
    RET
