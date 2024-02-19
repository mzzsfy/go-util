#include "textflag.h"

TEXT ·getG(SB), NOSPLIT, $0
    MOVW    g, ret+0(FP)
    RET
