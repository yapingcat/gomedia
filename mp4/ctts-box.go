package mp4

// aligned(8) class CompositionOffsetBox extends FullBox(‘ctts’, version = 0, 0) {
//     unsigned int(32) entry_count;
//     int i;
//     if (version==0) {
//         for (i=0; i < entry_count; i++) {
//             unsigned int(32) sample_count;
//             unsigned int(32) sample_offset;
//         }
//     }
//     else if (version == 1) {
//         for (i=0; i < entry_count; i++) {
//             unsigned int(32) sample_count;
//             signed int(32) sample_offset;
//         }
//     }
// }

type CompositionOffsetBox struct {
}
