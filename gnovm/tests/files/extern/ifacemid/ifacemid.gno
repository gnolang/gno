package ifacemid

import "filetests/extern/ifaceext"

// Mid embeds ifaceext.Sec, so its (flattened) method set is {ifaceext.sec};
// the unexported method's origin package stays ifaceext, not ifacemid.
type Mid interface{ ifaceext.Sec }
