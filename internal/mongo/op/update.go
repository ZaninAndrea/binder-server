package op

type Operator string

const (
	CurrentDate       Operator = "$currentDate"
	Inc               Operator = "$inc"
	Min               Operator = "$min"
	Max               Operator = "$max"
	Mul               Operator = "$mul"
	Rename            Operator = "$rename"
	Set               Operator = "$set"
	SetOnInsert       Operator = "$setOnInsert"
	Unset             Operator = "$unset"
	AddToSet          Operator = "$addToSet"
	Pop               Operator = "$pop"
	Pull              Operator = "$pull"
	Push              Operator = "$push"
	PullAll           Operator = "$pullAll"
	Each              Operator = "$each"
	Position          Operator = "$position"
	Sort              Operator = "$sort"
	Bit               Operator = "$bit"
	Explain           Operator = "$explain"
	Hint              Operator = "$hint"
	MaxTimeMS         Operator = "$maxTimeMS"
	OrderBy           Operator = "$orderby"
	Query             Operator = "$query"
	ReturnKey         Operator = "$returnKey"
	ShowDiskLoc       Operator = "$showDiskLoc"
	Natural           Operator = "$natural"
	Eq                Operator = "$eq"
	Gt                Operator = "$gt"
	Gte               Operator = "$gte"
	In                Operator = "$in"
	Lt                Operator = "$lt"
	Lte               Operator = "$lte"
	Ne                Operator = "$ne"
	Nin               Operator = "$nin"
	And               Operator = "$and"
	Not               Operator = "$not"
	Nor               Operator = "$nor"
	Or                Operator = "$or"
	Exists            Operator = "$exists"
	Type              Operator = "$type"
	Expr              Operator = "$expr"
	JSONSchema        Operator = "$jsonSchema"
	Mod               Operator = "$mod"
	Regex             Operator = "$regex"
	Text              Operator = "$text"
	Where             Operator = "$where"
	GeoIntersects     Operator = "$geoIntersects"
	GeoWithin         Operator = "$geoWithin"
	Near              Operator = "$near"
	NearSphere        Operator = "$nearSphere"
	All               Operator = "$all"
	ElemMatch         Operator = "$elemMatch"
	Size              Operator = "$size"
	BitsAllClear      Operator = "$bitsAllClear"
	BitsAllSet        Operator = "$bitsAllSet"
	BitsAnyClear      Operator = "$bitsAnyClear"
	BitsAnySet        Operator = "$bitsAnySet"
	Comment           Operator = "$comment"
	Dollar            Operator = "$"
	Meta              Operator = "$meta"
	Slice             Operator = "$slice"
	AddFields         Operator = "$addFields"
	Bucket            Operator = "$bucket"
	BucketAuto        Operator = "$bucketAuto"
	CollStats         Operator = "$collStats"
	Count             Operator = "$count"
	Facet             Operator = "$facet"
	GeoNear           Operator = "$geoNear"
	GraphLookup       Operator = "$graphLookup"
	Group             Operator = "$group"
	IndexStats        Operator = "$indexStats"
	Limit             Operator = "$limit"
	ListSessions      Operator = "$listSessions"
	Lookup            Operator = "$lookup"
	Match             Operator = "$match"
	Merge             Operator = "$merge"
	Out               Operator = "$out"
	PlanCacheStats    Operator = "$planCacheStats"
	Project           Operator = "$project"
	Redact            Operator = "$redact"
	ReplaceRoot       Operator = "$replaceRoot"
	ReplaceWith       Operator = "$replaceWith"
	Sample            Operator = "$sample"
	Skip              Operator = "$skip"
	SortByCount       Operator = "$sortByCount"
	Unwind            Operator = "$unwind"
	CurrentOp         Operator = "$currentOp"
	ListLocalSessions Operator = "$listLocalSessions"
	Abs               Operator = "$abs"
	Add               Operator = "$add"
	Ceil              Operator = "$ceil"
	Divide            Operator = "$divide"
	Exp               Operator = "$exp"
	Floor             Operator = "$floor"
	Ln                Operator = "$ln"
	Log               Operator = "$log"
	Log10             Operator = "$log10"
	Multiply          Operator = "$multiply"
	Pow               Operator = "$pow"
	Round             Operator = "$round"
	Sqrt              Operator = "$sqrt"
	Subtract          Operator = "$subtract"
	Trunc             Operator = "$trunc"
	ArrayToObject     Operator = "$arrayToObject"
	ConcatArrays      Operator = "$concatArrays"
	Filter            Operator = "$filter"
	IndexOfArray      Operator = "$indexOfArray"
	IsArray           Operator = "$isArray"
	Map               Operator = "$map"
	ObjectToArray     Operator = "$objectToArray"
	Range             Operator = "$range"
	Reduce            Operator = "$reduce"
	ReverseArray      Operator = "$reverseArray"
	Zip               Operator = "$zip"
	Cmp               Operator = "$cmp"
	Cond              Operator = "$cond"
	IfNull            Operator = "$ifNull"
	Switch            Operator = "$switch"
	DateFromParts     Operator = "$dateFromParts"
	DateFromString    Operator = "$dateFromString"
	DateToParts       Operator = "$dateToParts"
	DateToString      Operator = "$dateToString"
	DayOfMonth        Operator = "$dayOfMonth"
	DayOfWeek         Operator = "$dayOfWeek"
	DayOfYear         Operator = "$dayOfYear"
	Hour              Operator = "$hour"
	IsoDayOfWeek      Operator = "$isoDayOfWeek"
	IsoWeek           Operator = "$isoWeek"
	IsoWeekYear       Operator = "$isoWeekYear"
	Millisecond       Operator = "$millisecond"
	Minute            Operator = "$minute"
	Month             Operator = "$month"
	Second            Operator = "$second"
	ToDate            Operator = "$toDate"
	Week              Operator = "$week"
	Year              Operator = "$year"
	Literal           Operator = "$literal"
	MergeObjects      Operator = "$mergeObjects"
	AllElementsTrue   Operator = "$allElementsTrue"
	AnyElementTrue    Operator = "$anyElementTrue"
	SetDifference     Operator = "$setDifference"
	SetEquals         Operator = "$setEquals"
	SetIntersection   Operator = "$setIntersection"
	SetIsSubset       Operator = "$setIsSubset"
	SetUnion          Operator = "$setUnion"
	Concat            Operator = "$concat"
	IndexOfBytes      Operator = "$indexOfBytes"
	IndexOfCP         Operator = "$indexOfCP"
	Ltrim             Operator = "$ltrim"
	RegexFind         Operator = "$regexFind"
	RegexFindAll      Operator = "$regexFindAll"
	RegexMatch        Operator = "$regexMatch"
	Rtrim             Operator = "$rtrim"
	Split             Operator = "$split"
	StrLenBytes       Operator = "$strLenBytes"
	StrLenCP          Operator = "$strLenCP"
	Strcasecmp        Operator = "$strcasecmp"
	Substr            Operator = "$substr"
	SubstrBytes       Operator = "$substrBytes"
	SubstrCP          Operator = "$substrCP"
	ToLower           Operator = "$toLower"
	ToString          Operator = "$toString"
	Trim              Operator = "$trim"
	ToUpper           Operator = "$toUpper"
	Sin               Operator = "$sin"
	Cos               Operator = "$cos"
	Tan               Operator = "$tan"
	Asin              Operator = "$asin"
	Acos              Operator = "$acos"
	Atan              Operator = "$atan"
	Atan2             Operator = "$atan2"
	Asinh             Operator = "$asinh"
	Acosh             Operator = "$acosh"
	Atanh             Operator = "$atanh"
	DegreesToRadians  Operator = "$degreesToRadians"
	RadiansToDegrees  Operator = "$radiansToDegrees"
	Convert           Operator = "$convert"
	ToBool            Operator = "$toBool"
	ToDecimal         Operator = "$toDecimal"
	ToDouble          Operator = "$toDouble"
	ToInt             Operator = "$toInt"
	ToLong            Operator = "$toLong"
	ToObjectID        Operator = "$toObjectId"
	Avg               Operator = "$avg"
	First             Operator = "$first"
	Last              Operator = "$last"
	StdDevPop         Operator = "$stdDevPop"
	StdDevSamp        Operator = "$stdDevSamp"
	Sum               Operator = "$sum"
	Let               Operator = "$let"
)
