# Storage API





func NewStorage() *Storage
func (s *Storage) CreateEntityType(name string) (int, error)
func (s *Storage) CreateEntityTypeUnsafe(name string) (int, error)
func (s *Storage) CreateEntity(entity types.StorageEntity) (int, error) {
func (s *Storage) CreateEntityUnsafe(entity types.StorageEntity) (int, error) {
func (s *Storage) CreateEntityUniqueValue(entity types.StorageEntity) (int, bool, error) {
func (s *Storage) CreateEntityUniqueValueUnsafe(entity types.StorageEntity) (int, bool, error) {
func (s *Storage) GetEntityByPath(Type int, id int, context string) (types.StorageEntity, error) {
func (s *Storage) GetEntityByPathUnsafe(Type int, id int, context string) (types.StorageEntity, error) {
func (s *Storage) GetEntitiesByType(Type string, context string) (map[int]types.StorageEntity, error) {
func (s *Storage) GetEntitiesByTypeUnsafe(Type string, context string) (map[int]types.StorageEntity, error) {
func (s *Storage) GetEntitiesByValue(value string, mode string, context string) (map[int]types.StorageEntity, error) {
func (s *Storage) GetEntitiesByValueUnsafe(value string, mode string, context string) (map[int]types.StorageEntity, error) {
func (s *Storage) GetEntitiesByTypeAndValue(Type string, value string, mode string, context string) (map[int]types.StorageEntity, error) {
func (s *Storage) GetEntitiesByTypeAndValueUnsafe(Type string, value string, mode string, context string) (map[int]types.StorageEntity, error) {
func (s *Storage) UpdateEntity(entity types.StorageEntity) error {
func (s *Storage) UpdateEntityUnsafe(entity types.StorageEntity) error {
func (s *Storage) DeleteEntity(Type int, id int) {
func (s *Storage) DeleteEntityUnsafe(Type int, id int) {
func (s *Storage) GetRelation(srcType int, srcID int, targetType int, targetID int) (types.StorageRelation, error) {
func (s *Storage) GetRelationUnsafe(srcType int, srcID int, targetType int, targetID int) (types.StorageRelation, error) {
func (s *Storage) RelationExists(srcType int, srcID int, targetType int, targetID int) bool {
func (s *Storage) RelationExistsUnsafe(srcType int, srcID int, targetType int, targetID int) bool {
func (s *Storage) DeleteRelationList(relationList map[int]types.StorageRelation) {
func (s *Storage) DeleteRelationListUnsafe(relationList map[int]types.StorageRelation) {
func (s *Storage) DeleteRelation(sourceType int, sourceID int, targetType int, targetID int) {
func (s *Storage) DeleteRelationUnsafe(sourceType int, sourceID int, targetType int, targetID int) {
func (s *Storage) DeleteChildRelations(Type int, id int) error {
func (s *Storage) DeleteChildRelationsUnsafe(Type int, id int) error {
func (s *Storage) DeleteParentRelations(Type int, id int) error {
func (s *Storage) DeleteParentRelationsUnsafe(Type int, id int) error {
func (s *Storage) CreateRelation(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (bool, error) {
func (s *Storage) CreateRelationUnsafe(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (bool, error) {
func (s *Storage) UpdateRelation(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (types.StorageRelation, error) {
func (s *Storage) UpdateRelationUnsafe(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (types.StorageRelation, error) {
func (s *Storage) GetChildRelationsBySourceTypeAndSourceId(Type int, id int, context string) (map[int]types.StorageRelation, error) {
func (s *Storage) GetChildRelationsBySourceTypeAndSourceIdUnsafe(Type int, id int, context string) (map[int]types.StorageRelation, error) {
func (s *Storage) GetParentEntitiesByTargetTypeAndTargetIdAndSourceType(targetType int, targetID int, sourceType int, context string) map[int]types.StorageEntity {
func (s *Storage) GetParentEntitiesByTargetTypeAndTargetIdAndSourceTypeUnsafe(targetType int, targetID int, sourceType int, context string) map[int]types.StorageEntity {
func (s *Storage) GetParentRelationsByTargetTypeAndTargetId(targetType int, targetID int, context string) (map[int]types.StorageRelation, error) {
func (s *Storage) GetParentRelationsByTargetTypeAndTargetIdUnsafe(targetType int, targetID int, context string) (map[int]types.StorageRelation, error) {
func (s *Storage) GetEntityTypes() map[int]string {
func (s *Storage) GetEntityTypesUnsafe() map[int]string {
func (s *Storage) GetEntityRTypes() map[string]int {
func (s *Storage) GetEntityRTypesUnsafe() map[string]int {
func (s *Storage) TypeExists(strType string) bool {
func (s *Storage) TypeExistsUnsafe(strType string) bool {
func (s *Storage) EntityExists(Type int, id int) bool {
func (s *Storage) EntityExistsUnsafe(Type int, id int) bool {
func (s *Storage) TypeIdExists(id int) bool {
func (s *Storage) TypeIdExistsUnsafe(id int) bool {
func (s *Storage) GetTypeIdByString(strType string) (int, error) {
func (s *Storage) GetTypeIdByStringUnsafe(strType string) (int, error) {
func (s *Storage) GetTypeStringById(intType int) (string, error) {
func (s *Storage) GetTypeStringByIdUnsafe(intType int) (string, error) {
func (s *Storage) GetEntityAmount() int {
func (s *Storage) GetEntityAmountByType(intType int) (int, error) {
func (s *Storage) MapTransportData(data transport.TransportEntity) transport.TransportEntity {
func (s *Storage) GetEntitiesByQueryFilter(
func (s *Storage) GetEntitiesByQueryFilterAndSourceAddress(
func (s *Storage) BatchUpdateAddressList(addressList [][2]int, values map[string]string) {
func (s *Storage) BatchDeleteAddressList(addressList [][2]int) {
func (s *Storage) LinkAddressLists(from [][2]int, to [][2]int) int {
func (s *Storage) TraverseEnrich(entity *transport.TransportEntity, direction int, depth int) {