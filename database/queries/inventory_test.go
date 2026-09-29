package queries

// migratedQueries pins the complete bytes of the 21 SQL files copied from the
// legacy query client. These fixtures are independent of a sibling legacy checkout.
//
// Entries with fixed set were changed on purpose (Phase 4 and its follow-up), and
// their pin is the fixed bytes. The legacy originals stay in infrastructure/legacy.
// Every other entry still pins the bytes copied at extraction time. INSERT_image.sql also
// moved, from events/images/confirm/ to images/create/.
var migratedQueries = []struct {
	path   string
	size   int
	sha256 string
	fixed  bool
}{
	{"authorization/clubs/can_manage/IS_student_authorized_club.sql", 646, "c9ebad5086ec71c6d06d781e39b0586131a51d107b1229379e7bce14d2f60a9c", true},
	{"authorization/events/can_manage/IS_student_authorized_event.sql", 670, "e696c559c97d641c49156266cfb96eef9a67832327920a5498bb689bd6047f35", true},
	{"clubs/create/INSERT_club.sql", 54, "a934bef1c2ffc2a55491bd1bd75f89606ab322c6bb1e64494aff076c2fc2a57d", false},
	{"clubs/create/INSERT_club_info.sql", 78, "050e1bb658dcc126054bbfa39d5929e9f48ed0b564128b970d216dd02db2601a", false},
	{"clubs/events/list/SELECT_club_events.sql", 1841, "82798ec6ff36e2e0ee9b9c95c509711bf283a23f13ebbac78a9a5fa63b7390ad", true},
	{"clubs/get/SELECT_club.sql", 539, "ef6ca31de0eca939b258e8c7abfad25d070dc3d0849b91a8b2c0dbf3c3fc747b", true},
	{"clubs/list/SELECT_clubs.sql", 590, "2385091f4bf6b36080db0484ae3e9316bcae3d575c1ecca71d4443a27eae3f9f", true},
	{"clubs/members/create/INSERT_club_member.sql", 731, "3db2268fb21a4d99d8ada037c7ade713641ee78cc4ff62faff0c3b95e151fde6", true},
	{"clubs/members/leave/DELETE_club_member.sql", 809, "6a6a9616222cb04d0f2fcc68c43b212ece74dc9fb704d977d70d561e38b16745", true},
	{"events/create/INSERT_event.sql", 792, "8384888e2354e4aea211398588f8472c1d2d36118f1cae149d9da48012363bd3", true},
	{"events/create/INSERT_event_club_link.sql", 130, "2a039ddf9368d76e3eb0b70b862977542e266869f91f0ac2738c8042b2ff0591", false},
	{"events/create/INSERT_event_description.sql", 335, "742da857641ef7c307b5588fb56a3401180b6b2156059f35a29fbf4fe7bbedeb", true},
	{"events/images/confirm/INSERT_event_image.sql", 74, "7cfedca2b318385f7747228674b9828c4dbc985d7db445df7e0dea4c410245cd", false},
	{"images/create/INSERT_image.sql", 959, "99f3b079b861d64497d6814c51b4ab3956e6a93c8f3b58a10b50264d0cfa35fd", true},
	{"events/read/SELECT_events.sql", 1696, "cde6612a09386d80dac06de7007e720dd58015b447775791ce229233a623d684", true},
	{"me/clubs/list/SELECT_student_clubs.sql", 524, "8ac7d54c6257f6b19c25dfeaf0994b8bfa161b1d02d858a6621597f3969d6a6a", true},
	{"me/events/list/SELECT_student_events.sql", 1855, "de5f4edadee0d2870531ccb85ab34c4cb5a68fa68eee7e9519edfdf26c0741e8", true},
	{"me/get/SELECT_student_by_sub.sql", 189, "abe30d5eded05b5bea2162a1e93fcee66b20142bb47e5e0fac8b5833a96ddc47", false},
	{"students/ensure/EXISTS_student_by_sub.sql", 278, "1c36fc43596da415f30e3e15d88a2659fb02a22761209e88e0b4adc8e82a3a96", false},
	{"students/ensure/UPSERT_student.sql", 146, "b281775ddf5418ee7d6ed0b744b13740ba917b5072e84f6b01d95ad1462be9fb", false},
	{"students/ensure/UPSERT_student_sub_only.sql", 239, "1b5ec7e066fc3461115d756d2cf8d1c1cfe73e26b7a038574e7ff175b4b04d06", false},
}
